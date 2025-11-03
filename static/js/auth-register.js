/**
 * WebAuthn Registration Flow
 * Handles user account creation with passkey registration
 */

class AuthRegister {
    constructor() {
        // Progress tracking
        this.currentStep = 1;

        // Step 1: Account Details
        this.detailsForm = document.getElementById('registration-details-form');
        this.usernameInput = document.getElementById('reg-username');
        this.emailInput = document.getElementById('reg-email');
        this.termsCheckbox = document.getElementById('terms-accept');

        // Step 2: Passkey Creation
        this.passkeyButton = document.getElementById('create-passkey-btn');
        this.passkeyStatus = document.getElementById('passkey-status');
        this.backButton = document.getElementById('registration-back');

        // User data storage (temporary during registration)
        this.userData = {};

        this.init();
    }

    init() {
        // Step 1: Account details form
        this.detailsForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleAccountDetails();
        });

        // Real-time validation
        this.usernameInput.addEventListener('blur', () => {
            this.validateUsername();
        });

        this.emailInput.addEventListener('blur', () => {
            this.validateEmail();
        });

        // Step 2: Passkey creation
        this.passkeyButton.addEventListener('click', () => {
            this.handlePasskeyCreation();
        });

        this.backButton.addEventListener('click', () => {
            this.goToStep(1);
        });
    }

    async handleAccountDetails() {
        // Validate all fields
        const isUsernameValid = await this.validateUsername();
        const isEmailValid = await this.validateEmail();

        if (!isUsernameValid || !isEmailValid) {
            return;
        }

        if (!this.termsCheckbox.checked) {
            this.showError('You must accept the Terms of Service and Privacy Policy');
            return;
        }

        // Store user data
        this.userData = {
            username: this.usernameInput.value.trim(),
            email: this.emailInput.value.trim()
        };

        // Move to step 2
        this.goToStep(2);
    }

    async validateUsername() {
        const username = this.usernameInput.value.trim();
        const errorElement = document.getElementById('username-error');

        // Reset error state
        this.usernameInput.classList.remove('error');
        errorElement.hidden = true;

        // Check length and format
        if (username.length < 3 || username.length > 30) {
            this.showFieldError('username', 'Username must be 3-30 characters');
            return false;
        }

        if (!/^[a-zA-Z0-9_-]+$/.test(username)) {
            this.showFieldError('username', 'Username can only contain letters, numbers, underscores, and hyphens');
            return false;
        }

        // Check availability with server
        try {
            const response = await fetch('/auth/check-username', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username })
            });

            const result = await response.json();

            if (!result.available) {
                this.showFieldError('username', 'This username is already taken');
                return false;
            }

            return true;
        } catch (error) {
            console.error('Username validation error:', error);
            // Don't block on validation errors, server will catch it
            return true;
        }
    }

    async validateEmail() {
        const email = this.emailInput.value.trim();
        const errorElement = document.getElementById('email-error');

        // Reset error state
        this.emailInput.classList.remove('error');
        errorElement.hidden = true;

        // Basic email format check
        if (!email.includes('@') || !email.includes('.')) {
            this.showFieldError('email', 'Please enter a valid email address');
            return false;
        }

        // Check availability with server
        try {
            const response = await fetch('/auth/check-email', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email })
            });

            const result = await response.json();

            if (!result.available) {
                this.showFieldError('email', 'An account with this email already exists');
                return false;
            }

            return true;
        } catch (error) {
            console.error('Email validation error:', error);
            // Don't block on validation errors, server will catch it
            return true;
        }
    }

    async handlePasskeyCreation() {
        if (!this.isWebAuthnSupported()) {
            this.showError('Your browser does not support passkeys. Please use a modern browser.');
            return;
        }

        try {
            this.showPasskeyStatus('Preparing registration...');
            this.passkeyButton.disabled = true;
            this.backButton.disabled = true;

            // Step 1: Begin registration (get challenge from server)
            const beginResponse = await fetch('/auth/register/begin', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.userData)
            });

            if (!beginResponse.ok) {
                const error = await beginResponse.json();
                throw new Error(error.error || 'Failed to begin registration');
            }

            const options = await beginResponse.json();

            // Step 2: Convert options for WebAuthn API
            const publicKeyOptions = this.preparePublicKeyOptions(options);

            this.showPasskeyStatus('Create your passkey using your device...');

            // Step 3: Create credential with authenticator
            const credential = await navigator.credentials.create({
                publicKey: publicKeyOptions
            });

            this.showPasskeyStatus('Finalizing registration...');

            // Step 4: Send credential to server
            const finishResponse = await fetch('/auth/register/finish', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.credentialToJSON(credential))
            });

            if (!finishResponse.ok) {
                const error = await finishResponse.json();
                throw new Error(error.error || 'Registration failed');
            }

            // Success! Move to step 3
            this.goToStep(3);

        } catch (error) {
            console.error('WebAuthn registration error:', error);
            this.hidePasskeyStatus();
            this.passkeyButton.disabled = false;
            this.backButton.disabled = false;

            if (error.name === 'NotAllowedError') {
                this.showError('Registration was cancelled or timed out');
            } else if (error.name === 'InvalidStateError') {
                this.showError('A passkey already exists for this device');
            } else {
                this.showError(error.message || 'Failed to create passkey. Please try again.');
            }
        }
    }

    isWebAuthnSupported() {
        return window.PublicKeyCredential !== undefined &&
               navigator.credentials !== undefined &&
               navigator.credentials.create !== undefined;
    }

    preparePublicKeyOptions(options) {
        return {
            challenge: this.base64ToArrayBuffer(options.challenge),
            rp: options.rp,
            user: {
                id: this.base64ToArrayBuffer(options.user.id),
                name: options.user.name,
                displayName: options.user.displayName
            },
            pubKeyCredParams: options.pubKeyCredParams,
            timeout: options.timeout,
            authenticatorSelection: options.authenticatorSelection,
            attestation: options.attestation || 'none',
            excludeCredentials: options.excludeCredentials?.map(cred => ({
                type: cred.type,
                id: this.base64ToArrayBuffer(cred.id),
                transports: cred.transports
            })) || []
        };
    }

    credentialToJSON(credential) {
        return {
            id: credential.id,
            rawId: this.arrayBufferToBase64(credential.rawId),
            type: credential.type,
            response: {
                attestationObject: this.arrayBufferToBase64(credential.response.attestationObject),
                clientDataJSON: this.arrayBufferToBase64(credential.response.clientDataJSON)
            }
        };
    }

    goToStep(step) {
        // Update current step
        this.currentStep = step;

        // Hide all steps
        document.getElementById('registration-step-1').hidden = true;
        document.getElementById('registration-step-2').hidden = true;
        document.getElementById('registration-step-3').hidden = true;

        // Show current step
        document.getElementById(`registration-step-${step}`).hidden = false;

        // Update progress indicator
        document.querySelectorAll('.auth-progress-step').forEach((el, index) => {
            el.classList.remove('active', 'completed');
            if (index + 1 < step) {
                el.classList.add('completed');
            } else if (index + 1 === step) {
                el.classList.add('active');
            }
        });
    }

    showFieldError(field, message) {
        const input = document.getElementById(field === 'username' ? 'reg-username' : 'reg-email');
        const errorElement = document.getElementById(`${field}-error`);

        input.classList.add('error');
        errorElement.textContent = message;
        errorElement.hidden = false;
    }

    showPasskeyStatus(message) {
        this.passkeyStatus.querySelector('span').textContent = message;
        this.passkeyStatus.hidden = false;
    }

    hidePasskeyStatus() {
        this.passkeyStatus.hidden = true;
    }

    showError(message) {
        // Create or update error alert
        let alertDiv = document.querySelector('.auth-alert-error');
        if (!alertDiv) {
            alertDiv = document.createElement('div');
            alertDiv.className = 'auth-alert auth-alert-error';
            alertDiv.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" viewBox="0 0 256 256">
                    <path d="M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm-8,56a8,8,0,0,1,16,0v56a8,8,0,0,1-16,0Zm8,104a12,12,0,1,1,12-12A12,12,0,0,1,128,184Z"></path>
                </svg>
                <span></span>
            `;
            const authCard = document.querySelector('.auth-card');
            authCard.insertBefore(alertDiv, authCard.querySelector('.auth-progress'));
        }
        alertDiv.querySelector('span').textContent = message;
        alertDiv.hidden = false;

        // Scroll to top to show error
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }

    // Utility functions for base64 <-> ArrayBuffer conversion
    base64ToArrayBuffer(base64) {
        // Handle URL-safe base64
        base64 = base64.replace(/-/g, '+').replace(/_/g, '/');
        const binaryString = window.atob(base64);
        const bytes = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
            bytes[i] = binaryString.charCodeAt(i);
        }
        return bytes.buffer;
    }

    arrayBufferToBase64(buffer) {
        const bytes = new Uint8Array(buffer);
        let binary = '';
        for (let i = 0; i < bytes.byteLength; i++) {
            binary += String.fromCharCode(bytes[i]);
        }
        // Return URL-safe base64
        return window.btoa(binary)
            .replace(/\+/g, '-')
            .replace(/\//g, '_')
            .replace(/=/g, '');
    }
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    new AuthRegister();
});
