// my-theme/login/resources/js/redirect-error.js
window.addEventListener('DOMContentLoaded', () => {
    const params = new URLSearchParams(window.location.search);
        // Указываем ваш frontend URL
        const frontendUrl = 'https://localhost:3000/oauth/error';
        const error = encodeURIComponent(params.get('error'));
        const description = encodeURIComponent(params.get('error_description') || '');
        window.location.href = `${frontendUrl}?error=${error}&description=${description}`;
});