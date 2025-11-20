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
            const postCard = button.closest('.post-card');
            const postID = postCard.dataset.postId;
            const commentsSection = document.getElementById(`comments-${postID}`);

            if (!commentsSection) return;

            // Toggle visibility
            const isHidden = commentsSection.style.display === 'none';

            if (isHidden) {
                commentsSection.style.display = 'block';

                // Load comments if not already loaded
                if (!commentsSection.dataset.loaded) {
                    fetch(`/api/posts/${postID}/comments`)
                        .then(response => response.json())
                        .then(comments => {
                            // Render comments (simplified - ideally use HTMX)
                            commentsSection.innerHTML = `
                                <div class="comment-list-container">
                                    <div class="comment-list">
                                        ${comments.length > 0 ? comments.map(c => renderComment(c)).join('') : '<div class="comment-list-empty"><svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" fill="currentColor" viewBox="0 0 256 256"><path d="M128,24A104,104,0,0,0,36.18,176.88L24.83,210.93a16,16,0,0,0,20.24,20.24l34.05-11.35A104,104,0,1,0,128,24Z"></path></svg><p>No comments yet</p><span>Be the first to share your thoughts</span></div>'}
                                    </div>
                                    ${renderCommentForm(postID)}
                                </div>
                            `;
                            commentsSection.dataset.loaded = 'true';
                            initializeCommentForm(postID);
                        })
                        .catch(error => {
                            console.error('Error loading comments:', error);
                            commentsSection.innerHTML = '<p class="error">Failed to load comments</p>';
                        });
                }
            } else {
                commentsSection.style.display = 'none';
            }
            return;
        }
    });
})();

// Comment rendering helpers
function renderComment(comment) {
    const avatarHTML = comment.author_avatar
        ? `<img src="${comment.author_avatar}" alt="${comment.author_name}" />`
        : `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="currentColor" viewBox="0 0 256 256"><path d="M128,80a48,48,0,1,0,48,48A48.05,48.05,0,0,0,128,80Zm0,80a32,32,0,1,1,32-32A32,32,0,0,1,128,160Zm88-120H40A16,16,0,0,0,24,56V200a16,16,0,0,0,16,16H216a16,16,0,0,0,16-16V56A16,16,0,0,0,216,40Z"></path></svg>`;

    return `
        <div class="comment-card" data-comment-id="${comment.id}" data-post-id="${comment.post_id}">
            <div class="comment-header">
                <div class="comment-author">
                    <div class="comment-avatar">${avatarHTML}</div>
                    <div class="comment-author-info">
                        <a href="/profile/${comment.author_profile_id}" class="comment-author-name">${comment.author_name}</a>
                        <div class="comment-meta">
                            <time datetime="${comment.created_at}" class="comment-time">${comment.formatted_time || 'just now'}</time>
                            ${comment.is_edited ? '<span class="comment-edited">• edited</span>' : ''}
                        </div>
                    </div>
                </div>
            </div>
            <div class="comment-body">
                <div class="comment-plaintext">${escapeHTML(comment.body)}</div>
            </div>
        </div>
    `;
}

function renderCommentForm(postID) {
    return `
        <form class="comment-form" data-post-id="${postID}">
            <div class="comment-form-input">
                <textarea name="body" placeholder="Write a comment..." rows="2" maxlength="2000" required></textarea>
                <div class="comment-form-footer">
                    <div class="comment-form-hint">
                        <span class="char-count"><span class="char-current">0</span>/2000</span>
                    </div>
                    <button type="submit" class="btn-primary btn-sm" disabled>Comment</button>
                </div>
            </div>
        </form>
    `;
}

function initializeCommentForm(postID) {
    const form = document.querySelector(`.comment-form[data-post-id="${postID}"]`);
    if (!form) return;

    const textarea = form.querySelector('textarea');
    const submitBtn = form.querySelector('button[type="submit"]');
    const charCurrent = form.querySelector('.char-current');

    textarea.addEventListener('input', function() {
        const length = this.value.trim().length;
        charCurrent.textContent = this.value.length;
        submitBtn.disabled = length === 0;
    });

    form.addEventListener('submit', function(e) {
        e.preventDefault();

        const body = textarea.value.trim();
        if (!body) return;

        submitBtn.disabled = true;
        submitBtn.textContent = 'Posting...';

        fetch(`/api/posts/${postID}/comments`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ body: body, body_format: 'plaintext' })
        })
        .then(response => {
            if (!response.ok) throw new Error('Failed to post comment');
            return response.json();
        })
        .then(comment => {
            // Add new comment to list
            const commentList = form.previousElementSibling;
            const emptyState = commentList.querySelector('.comment-list-empty');

            if (emptyState) {
                emptyState.remove();
            }

            commentList.insertAdjacentHTML('beforeend', renderComment(comment));

            // Clear form
            textarea.value = '';
            charCurrent.textContent = '0';
            submitBtn.disabled = true;
            submitBtn.textContent = 'Comment';

            // Update comment count
            const postCard = form.closest('.post-card');
            const commentBtn = postCard.querySelector('.engagement-btn[data-action="comment"] span');
            if (commentBtn) {
                const currentCount = parseInt(commentBtn.textContent) || 0;
                const newCount = currentCount + 1;
                commentBtn.textContent = `${newCount} ${newCount === 1 ? 'comment' : 'comments'}`;
            }
        })
        .catch(error => {
            console.error('Error posting comment:', error);
            alert('Failed to post comment. Please try again.');
            submitBtn.disabled = false;
            submitBtn.textContent = 'Comment';
        });
    });
}

function escapeHTML(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}
