/**
 * HTMX Helpers
 * Utilities for managing modals, toasts, and HTMX interactions
 */

// Toast management
const ToastManager = {
  container: null,

  init() {
    // Create toast container if it doesn't exist
    if (!document.getElementById('toast-container')) {
      this.container = document.createElement('div');
      this.container.id = 'toast-container';
      document.body.appendChild(this.container);
    } else {
      this.container = document.getElementById('toast-container');
    }
  },

  show(type, message, duration = 4000) {
    if (!this.container) this.init();

    // Create toast HTML
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.setAttribute('role', 'alert');
    toast.setAttribute('aria-live', 'polite');

    // Icon based on type
    let iconPath = '';
    if (type === 'success') {
      iconPath = 'M173.66,98.34a8,8,0,0,1,0,11.32l-56,56a8,8,0,0,1-11.32,0l-24-24a8,8,0,0,1,11.32-11.32L112,148.69l50.34-50.35A8,8,0,0,1,173.66,98.34ZM232,128A104,104,0,1,1,128,24,104.11,104.11,0,0,1,232,128Zm-16,0a88,88,0,1,0-88,88A88.1,88.1,0,0,0,216,128Z';
    } else if (type === 'error') {
      iconPath = 'M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm0,192a88,88,0,1,1,88-88A88.1,88.1,0,0,1,128,216Zm-8-80V80a8,8,0,0,1,16,0v56a8,8,0,0,1-16,0Zm20,36a12,12,0,1,1-12-12A12,12,0,0,1,140,172Z';
    } else {
      iconPath = 'M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm0,192a88,88,0,1,1,88-88A88.1,88.1,0,0,1,128,216Zm16-40a8,8,0,0,1-8,8,16,16,0,0,1-16-16V128a8,8,0,0,1,0-16,16,16,0,0,1,16,16v40A8,8,0,0,1,144,176ZM112,84a12,12,0,1,1,12,12A12,12,0,0,1,112,84Z';
    }

    toast.innerHTML = `
      <svg class="toast-icon" width="20" height="20" viewBox="0 0 256 256" fill="currentColor">
        <path d="${iconPath}"></path>
      </svg>
      <span class="toast-message">${message}</span>
      <button class="toast-close" aria-label="Dismiss notification">
        <svg width="16" height="16" viewBox="0 0 256 256" fill="currentColor">
          <path d="M205.66,194.34a8,8,0,0,1-11.32,11.32L128,139.31,61.66,205.66a8,8,0,0,1-11.32-11.32L116.69,128,50.34,61.66A8,8,0,0,1,61.66,50.34L128,116.69l66.34-66.35a8,8,0,0,1,11.32,11.32L139.31,128Z"></path>
        </svg>
      </button>
    `;

    // Add close handler
    const closeBtn = toast.querySelector('.toast-close');
    closeBtn.addEventListener('click', () => this.remove(toast));

    // Add to container
    this.container.appendChild(toast);

    // Auto-dismiss after duration
    if (duration > 0) {
      setTimeout(() => this.remove(toast), duration);
    }

    return toast;
  },

  remove(toast) {
    toast.classList.add('removing');
    setTimeout(() => {
      if (toast.parentElement) {
        toast.parentElement.removeChild(toast);
      }
    }, 200);
  }
};

// Modal management
const ModalManager = {
  modal: null,
  previousFocus: null,

  init() {
    this.modal = document.getElementById('modal');
    if (!this.modal) {
      console.warn('Modal element not found');
      return;
    }

    // Close on Escape key
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && this.isOpen()) {
        this.close();
      }
    });

    // Trap focus in modal
    this.modal.addEventListener('keydown', (e) => {
      if (e.key === 'Tab' && this.isOpen()) {
        this.trapFocus(e);
      }
    });
  },

  open(content) {
    if (!this.modal) this.init();

    // Store current focus
    this.previousFocus = document.activeElement;

    // Set content
    if (content) {
      this.modal.innerHTML = content;
    }

    // Show modal
    this.modal.classList.add('active');
    document.body.style.overflow = 'hidden';

    // Focus first focusable element
    setTimeout(() => this.focusFirstElement(), 10);
  },

  close() {
    if (!this.modal) return;

    this.modal.classList.remove('active');
    document.body.style.overflow = '';

    // Restore focus
    if (this.previousFocus) {
      this.previousFocus.focus();
      this.previousFocus = null;
    }

    // Clear content after animation
    setTimeout(() => {
      if (!this.isOpen()) {
        this.modal.innerHTML = '';
      }
    }, 200);
  },

  isOpen() {
    return this.modal && this.modal.classList.contains('active');
  },

  focusFirstElement() {
    if (!this.modal) return;

    const focusable = this.modal.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );

    if (focusable.length > 0) {
      focusable[0].focus();
    }
  },

  trapFocus(e) {
    const focusable = this.modal.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );

    const firstFocusable = focusable[0];
    const lastFocusable = focusable[focusable.length - 1];

    if (e.shiftKey) {
      if (document.activeElement === firstFocusable) {
        lastFocusable.focus();
        e.preventDefault();
      }
    } else {
      if (document.activeElement === lastFocusable) {
        firstFocusable.focus();
        e.preventDefault();
      }
    }
  }
};

// Global function for closing modal (used in templates)
function closeModal() {
  ModalManager.close();
}

// Initialize on DOM ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    ToastManager.init();
    ModalManager.init();
  });
} else {
  ToastManager.init();
  ModalManager.init();
}

// HTMX event listeners
document.body.addEventListener('htmx:afterSwap', function(evt) {
  // If swapping into modal, open it
  if (evt.detail.target.id === 'modal') {
    ModalManager.open();
  }

  // Handle toast notifications in response
  const toasts = evt.detail.target.querySelectorAll('.toast');
  toasts.forEach(toast => {
    // Move toast to toast container
    const type = toast.classList.contains('toast-success') ? 'success' :
                 toast.classList.contains('toast-error') ? 'error' : 'info';
    const message = toast.querySelector('.toast-message').textContent;
    ToastManager.show(type, message);
    toast.remove();
  });
});

document.body.addEventListener('htmx:beforeRequest', function(evt) {
  // Add loading state to buttons
  const elt = evt.detail.elt;
  if (elt.tagName === 'BUTTON' || elt.tagName === 'FORM') {
    elt.classList.add('htmx-request');
  }
});

document.body.addEventListener('htmx:afterRequest', function(evt) {
  // Remove loading state
  const elt = evt.detail.elt;
  if (elt.tagName === 'BUTTON' || elt.tagName === 'FORM') {
    elt.classList.remove('htmx-request');
  }

  // Handle errors
  if (!evt.detail.successful) {
    ToastManager.show('error', 'Something went wrong. Please try again.');
  }
});

document.body.addEventListener('htmx:responseError', function(evt) {
  const xhr = evt.detail.xhr;
  let message = 'Network error. Please try again.';

  if (xhr.status === 400) {
    message = 'Invalid input. Please check your information.';
  } else if (xhr.status === 403) {
    message = 'You don\'t have permission to do that.';
  } else if (xhr.status === 404) {
    message = 'Not found.';
  } else if (xhr.status >= 500) {
    message = 'Server error. Please try again later.';
  }

  ToastManager.show('error', message);
});

// Export for use in other scripts
window.ToastManager = ToastManager;
window.ModalManager = ModalManager;
