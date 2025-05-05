document.addEventListener('DOMContentLoaded', function() {
    // Post filter functionality
    const categoryFilter = document.getElementById('category-filter');
    if (categoryFilter) {
        categoryFilter.addEventListener('change', function() {
            const categoryId = this.value;
            if (categoryId) {
                window.location.href = `/?category=${categoryId}`;
            } else {
                window.location.href = '/';
            }
        });
    }

    // Add smooth scrolling to comment section
    const commentLinks = document.querySelectorAll('.comment-link');
    commentLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            const commentSection = document.querySelector(this.getAttribute('href'));
            if (commentSection) {
                commentSection.scrollIntoView({ behavior: 'smooth' });
                
                // Focus on comment input
                const commentInput = document.getElementById('comment-input');
                if (commentInput) {
                    commentInput.focus();
                }
            }
        });
    });

    // Preview comment
    const commentInput = document.getElementById('comment-input');
    const commentPreview = document.getElementById('comment-preview');
    const previewToggle = document.getElementById('preview-toggle');
    
    if (commentInput && commentPreview && previewToggle) {
        previewToggle.addEventListener('click', function(e) {
            e.preventDefault();
            const isPreviewVisible = commentPreview.style.display === 'block';
            
            if (isPreviewVisible) {
                commentPreview.style.display = 'none';
                commentInput.style.display = 'block';
                this.textContent = 'Preview';
            } else {
                commentPreview.innerHTML = commentInput.value;
                commentPreview.style.display = 'block';
                commentInput.style.display = 'none';
                this.textContent = 'Edit';
            }
        });
    }

    // Form validation
    const validateForm = (formId, rules) => {
        const form = document.getElementById(formId);
        if (!form) return;

        form.addEventListener('submit', function(e) {
            let isValid = true;
            const errorMessages = [];

            // Check each validation rule
            for (const field in rules) {
                const input = form.querySelector(`[name="${field}"]`);
                if (!input) continue;

                const value = input.value.trim();
                const fieldRules = rules[field];

                if (fieldRules.required && value === '') {
                    isValid = false;
                    errorMessages.push(`${fieldRules.label} is required`);
                    input.classList.add('error');
                } else if (fieldRules.minLength && value.length < fieldRules.minLength) {
                    isValid = false;
                    errorMessages.push(`${fieldRules.label} must be at least ${fieldRules.minLength} characters`);
                    input.classList.add('error');
                } else if (fieldRules.match && input.value !== form.querySelector(`[name="${fieldRules.match}"]`).value) {
                    isValid = false;
                    errorMessages.push(`${fieldRules.label} does not match`);
                    input.classList.add('error');
                } else if (fieldRules.email && !isValidEmail(value)) {
                    isValid = false;
                    errorMessages.push(`${fieldRules.label} must be a valid email address`);
                    input.classList.add('error');
                } else {
                    input.classList.remove('error');
                }
            }

            if (!isValid) {
                e.preventDefault();
                displayErrors(formId, errorMessages);
            }
        });
    };

    // Email validation helper
    const isValidEmail = (email) => {
        const re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
        return re.test(String(email).toLowerCase());
    };

    // Display form errors
    const displayErrors = (formId, errors) => {
        const form = document.getElementById(formId);
        if (!form) return;

        let errorContainer = form.querySelector('.error-messages');
        if (!errorContainer) {
            errorContainer = document.createElement('div');
            errorContainer.className = 'error-messages';
            form.insertBefore(errorContainer, form.firstChild);
        }

        errorContainer.innerHTML = '';
        if (errors.length > 0) {
            const ul = document.createElement('ul');
            errors.forEach(error => {
                const li = document.createElement('li');
                li.textContent = error;
                ul.appendChild(li);
            });
            errorContainer.appendChild(ul);
        }
    };

    // Set up validation for register form
    validateForm('register-form', {
        username: { 
            required: true, 
            minLength: 3,
            label: 'Username'
        },
        email: { 
            required: true, 
            email: true,
            label: 'Email'
        },
        password: { 
            required: true, 
            minLength: 6,
            label: 'Password'
        },
        confirm_password: { 
            required: true, 
            match: 'password',
            label: 'Confirm password'
        }
    });

    // Set up validation for login form
    validateForm('login-form', {
        email: { 
            required: true, 
            email: true,
            label: 'Email'
        },
        password: { 
            required: true,
            label: 'Password'
        }
    });

    // Set up validation for post creation form
    validateForm('post-form', {
        title: { 
            required: true,
            minLength: 5,
            label: 'Title'
        },
        content: { 
            required: true,
            minLength: 10,
            label: 'Content'
        }
    });

    // Set up validation for comment form
    validateForm('comment-form', {
        content: { 
            required: true,
            minLength: 3,
            label: 'Comment'
        }
    });

    // Handle category selection in post form
    const categoryCheckboxes = document.querySelectorAll('.category-checkbox');
    if (categoryCheckboxes.length > 0) {
        const categoryError = document.getElementById('category-error');
        
        // Monitor changes to update validation status
        categoryCheckboxes.forEach(checkbox => {
            checkbox.addEventListener('change', function() {
                const anyChecked = Array.from(categoryCheckboxes).some(cb => cb.checked);
                if (categoryError) {
                    categoryError.style.display = anyChecked ? 'none' : 'block';
                }
            });
        });
        
        // Add form submission validation for categories
        const postForm = document.getElementById('post-form');
        if (postForm) {
            postForm.addEventListener('submit', function(e) {
                const anyChecked = Array.from(categoryCheckboxes).some(cb => cb.checked);
                if (!anyChecked) {
                    e.preventDefault();
                    if (categoryError) {
                        categoryError.style.display = 'block';
                    }
                }
            });
        }
    }
});
