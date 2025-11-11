/**
 * WebAuthn Login Flow
 * Handles passwordless authentication using WebAuthn
 */

class AuthLogin {
    constructor() {
        this.form = document.getElementById('webauthn-login-form');
        this.usernameInput = document.getElementById('username');
        this.statusElement = document.getElementById('webauthn-status');
        this.emailToggle = document.getElementById('email-login-toggle');
        this.emailForm = document.getElementById('email-login-form');
        this.emailCancel = document.getElementById('email-login-cancel');

        this.init();
    }

    init() {
        // Check WebAuthn support
        if (!this.isWebAuthnSupported()) {
            this.showEmailFallback();
            return;
        }

        // WebAuthn login form
        this.form.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleWebAuthnLogin();
        });

        // Email fallback toggle
        this.emailToggle.addEventListener('click', () => {
            this.showEmailForm();
        });

        this.emailCancel.addEventListener('click', () => {
            this.hideEmailForm();
        });

        // Email login form
        this.emailForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleEmailLogin();
        });
    }

    isWebAuthnSupported() {
        return window.PublicKeyCredential !== undefined &&
               navigator.credentials !== undefined &&
               navigator.credentials.create !== undefined;
    }

    async handleWebAuthnLogin() {
        const username = this.usernameInput.value.trim();

        if (!username) {
            this.showError('Please enter your username or email');
            return;
        }

        try {
            this.showStatus('Preparing authentication...');
            this.disableForm();

            // Step 1: Begin authentication (get challenge from server)
            const beginResponse = await fetch('/auth/login/begin', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username })
            });

            if (!beginResponse.ok) {
                const error = await beginResponse.json();
                throw new Error(error.error || 'Failed to begin authentication');
            }

            const options = await beginResponse.json();

            // Step 2: Convert challenge and credential IDs from base64
            const publicKeyOptions = this.preparePublicKeyOptions(options);

            this.showStatus('Waiting for your passkey...');

            // Step 3: Get credential from authenticator
            const credential = await navigator.credentials.get({
                publicKey: publicKeyOptions
            });

            this.showStatus('Verifying...');

            // Step 4: Send credential to server for verification
            const finishResponse = await fetch('/auth/login/finish', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(this.credentialToJSON(credential))
            });

            if (!finishResponse.ok) {
                const error = await finishResponse.json();
                throw new Error(error.error || 'Authentication failed');
            }

            const result = await finishResponse.json();

            // Success! Redirect to circles
            this.showStatus('Success! Redirecting...');
            setTimeout(() => {
                window.location.href = result.redirect || '/circles';
            }, 500);

        } catch (error) {
            console.error('WebAuthn login error:', error);
            this.hideStatus();
            this.enableForm();

            if (error.name === 'NotAllowedError') {
                this.showError('Authentication was cancelled or timed out');
            } else if (error.name === 'InvalidStateError') {
                this.showError('This passkey is not registered');
            } else {
                this.showError(error.message || 'Authentication failed. Please try again.');
            }
        }
    }

    preparePublicKeyOptions(options) {
        return {
            challenge: this.base64ToArrayBuffer(options.challenge),
            timeout: options.timeout,
            rpId: options.rpId,
            allowCredentials: options.allowCredentials?.map(cred => ({
                type: cred.type,
                id: this.base64ToArrayBuffer(cred.id),
                transports: cred.transports
            })) || [],
            userVerification: options.userVerification || 'preferred'
        };
    }

    credentialToJSON(credential) {
        return {
            id: credential.id,
            rawId: this.arrayBufferToBase64(credential.rawId),
            type: credential.type,
            response: {
                authenticatorData: this.arrayBufferToBase64(credential.response.authenticatorData),
                clientDataJSON: this.arrayBufferToBase64(credential.response.clientDataJSON),
                signature: this.arrayBufferToBase64(credential.response.signature),
                userHandle: credential.response.userHandle ?
                    this.arrayBufferToBase64(credential.response.userHandle) : null
            }
        };
    }

    async handleEmailLogin() {
        const email = document.getElementById('email').value.trim();

        if (!email) {
            this.showError('Please enter your email address');
            return;
        }

        try {
            const response = await fetch('/auth/email/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error || 'Failed to send magic link');
            }

            // Show success message
            this.showSuccess('Check your email! We sent you a magic link to sign in.');
            this.emailForm.reset();

        } catch (error) {
            console.error('Email login error:', error);
            this.showError(error.message || 'Failed to send magic link. Please try again.');
        }
    }

    showEmailFallback() {
        document.getElementById('webauthn-login').style.display = 'none';
        this.emailForm.hidden = false;
        this.emailToggle.style.display = 'none';
    }

    showEmailForm() {
        this.emailForm.hidden = false;
        this.emailToggle.hidden = true;
    }

    hideEmailForm() {
        this.emailForm.hidden = true;
        this.emailToggle.hidden = false;
    }

    showStatus(message) {
        this.statusElement.querySelector('span').textContent = message;
        this.statusElement.hidden = false;
    }

    hideStatus() {
        this.statusElement.hidden = true;
    }

    disableForm() {
        this.form.querySelectorAll('button, input').forEach(el => {
            el.disabled = true;
        });
    }

    enableForm() {
        this.form.querySelectorAll('button, input').forEach(el => {
            el.disabled = false;
        });
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
            document.querySelector('.auth-card').insertBefore(
                alertDiv,
                document.querySelector('.auth-section')
            );
        }
        alertDiv.querySelector('span').textContent = message;
        alertDiv.hidden = false;
    }

    showSuccess(message) {
        // Create or update success alert
        let alertDiv = document.querySelector('.auth-alert-success');
        if (!alertDiv) {
            alertDiv = document.createElement('div');
            alertDiv.className = 'auth-alert auth-alert-success';
            alertDiv.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" viewBox="0 0 256 256">
                    <path d="M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm45.66,85.66-56,56a8,8,0,0,1-11.32,0l-24-24a8,8,0,0,1,11.32-11.32L112,148.69l50.34-50.35a8,8,0,0,1,11.32,11.32Z"></path>
                </svg>
                <span></span>
            `;
            document.querySelector('.auth-card').insertBefore(
                alertDiv,
                document.querySelector('.auth-section')
            );
        }
        alertDiv.querySelector('span').textContent = message;
        alertDiv.hidden = false;
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
    new AuthLogin();
});
