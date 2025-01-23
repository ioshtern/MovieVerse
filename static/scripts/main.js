// scripts/main.js

// Helper function to get a cookie by name
function getCookie(name) {
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [key, value] = cookie.split('=').map(c => c.trim());
        if (key === name) return value;
    }
    return null;
}

// Show/Hide buttons based on user login state
function updateNavButtons() {
    const loginButton = document.querySelector('a[href="login.html"]');
    const signupButton = document.querySelector('a[href="signup.html"]');
    const logoutButton = document.createElement('button');

    // Check if userToken cookie exists
    const userToken = getCookie('userToken');
    if (userToken) {
        // User is logged in
        loginButton.style.display = 'none';
        signupButton.style.display = 'none';

        // Add Logout button
        logoutButton.textContent = 'Logout';
        logoutButton.classList.add('btn', 'btn-outline-light', 'ms-2');
        logoutButton.addEventListener('click', () => {
            // Remove the cookie and reload the page
            document.cookie = 'userToken=; Max-Age=0; path=/';
            window.location.reload();
        });
        loginButton.parentNode.appendChild(logoutButton);
    } else {
        // User is not logged in
        loginButton.style.display = '';
        signupButton.style.display = '';
        if (logoutButton.parentNode) {
            logoutButton.parentNode.removeChild(logoutButton);
        }
    }
}

// Run on page load
document.addEventListener('DOMContentLoaded', updateNavButtons);
