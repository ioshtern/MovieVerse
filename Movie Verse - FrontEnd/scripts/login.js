document.querySelector('#loginButton').addEventListener('click', function (event) {
    event.preventDefault(); 

    var emailField = document.querySelector('#email');
    var passwordField = document.querySelector('#password');
    var email = emailField.value.trim();
    var password = passwordField.value;

    function showError(field, message) {
        field.style.borderColor = 'red';
        var errorElement = field.nextElementSibling;
        if (!errorElement || !errorElement.classList.contains('error-message')) {
            errorElement = document.createElement('div');
            errorElement.classList.add('error-message');
            errorElement.style.color = 'red';
            errorElement.style.fontSize = '0.9em';
            field.insertAdjacentElement('afterend', errorElement);
        }
        errorElement.textContent = message;
    }

    function clearError(field) {
        field.style.borderColor = '';
        var errorElement = field.nextElementSibling;
        if (errorElement && errorElement.classList.contains('error-message')) {
            errorElement.remove();
        }
    }

    clearError(emailField);
    clearError(passwordField);

    if (email === 'sunnatovdavlat060605@gmail.com' && password === 'davlat123') {
        window.location.href = 'admin.html';
        return;
    }

    var users = JSON.parse(localStorage.getItem('users')) || [];
    var user = users.find(function (user) {
        return user.email === email && user.password === password;
    });

    if (user) {
        localStorage.setItem('loggedInUser', JSON.stringify(user));
        window.location.href = 'index.html';
    } else {
        if (!email) {
            showError(emailField, 'Email cannot be empty.');
        } else {
            showError(emailField, 'Invalid email or password.');
            showError(passwordField, 'Invalid email or password.');

        }
        if (!password) {
            showError(passwordField, 'Password cannot be empty.');
        }
    }
});
