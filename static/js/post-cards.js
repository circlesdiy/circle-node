// Post card event delegation
// Handles all interactions for post cards using event delegation for performance
(function() {
    'use strict';

    // Close all dropdowns when clicking outside
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.post-actions-menu')) {
            document.querySelectorAll('.action-dropdown').forEach(d => {
                d.hidden = true;
            });
        }
    });

    // Delegate all post card events to the document
    document.addEventListener('click', async (e) => {
        const target = e.target;
        const button = target.closest('button');

        if (!button) return;

        // Find the parent post card
        const postCard = button.closest('.post-card');
        if (!postCard) return;

        const postID = postCard.dataset.postId;

        // Handle content warning toggle
        if (button.hasAttribute('data-cw-toggle')) {
            const cwContent = postCard.querySelector('[data-cw-content]');
            if (cwContent) {
                cwContent.classList.toggle('content-hidden');
                button.textContent = cwContent.classList.contains('content-hidden')
                    ? 'Show Content'
                    : 'Hide Content';
            }
            return;
        }

        // Handle dropdown menu toggle
        if (button.classList.contains('post-action-btn')) {
            e.stopPropagation();
            const dropdown = postCard.querySelector('.action-dropdown');
            if (dropdown) {
                // Close all other dropdowns first
                document.querySelectorAll('.action-dropdown').forEach(d => {
                    if (d !== dropdown) d.hidden = true;
                });
                dropdown.hidden = !dropdown.hidden;
            }
            return;
        }

        // Handle dropdown actions (edit/delete)
        if (button.classList.contains('dropdown-action')) {
            e.stopPropagation();
            const action = button.dataset.action;
            const dropdown = postCard.querySelector('.action-dropdown');

            if (action === 'delete') {
                if (confirm('Are you sure you want to delete this post?')) {
                    try {
                        const response = await fetch(`/api/posts/${postID}`, {
                            method: 'DELETE',
                        });

                        if (response.ok) {
                            postCard.remove();
                        } else {
                            alert('Failed to delete post');
                        }
                    } catch (error) {
                        console.error('Error deleting post:', error);
                        alert('Failed to delete post');
                    }
                }
            } else if (action === 'edit') {
                // TODO: Implement edit functionality
                alert('Edit functionality coming soon');
            }

            if (dropdown) {
                dropdown.hidden = true;
            }
            return;
        }

        // Handle like/reaction button
        if (button.classList.contains('reaction-btn') && button.dataset.action === 'like') {
            const isActive = button.dataset.active === 'true';
            const countSpan = button.querySelector('.reaction-count');
            let count = parseInt(countSpan.textContent) || 0;

            try {
                if (isActive) {
                    // Remove reaction
                    const response = await fetch(`/api/posts/${postID}/reactions`, {
                        method: 'DELETE',
                    });

                    if (response.ok) {
                        button.dataset.active = 'false';
                        countSpan.textContent = Math.max(0, count - 1);
                    }
                } else {
                    // Add reaction
                    const response = await fetch(`/api/posts/${postID}/reactions`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({ key: 'like' }),
                    });

                    if (response.ok) {
                        button.dataset.active = 'true';
                        countSpan.textContent = count + 1;
                    }
                }
            } catch (error) {
                console.error('Error toggling reaction:', error);
            }
            return;
        }

        // Handle comment button
        if (button.classList.contains('engagement-btn') && button.dataset.action === 'comment') {
            // TODO: Implement comment functionality
            alert('Comment functionality coming soon');
            return;
        }
    });
})();
