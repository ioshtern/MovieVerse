document.querySelector('#signupButton').addEventListener('click', function (event) {
    event.preventDefault();

    var usernameField = document.querySelector('#username');
    var emailField = document.querySelector('#email');
    var passwordField = document.querySelector('#password');
    var username = usernameField.value.trim();
    var email = emailField.value.trim();
    var password = passwordField.value;

    // Function to display error message and highlight field
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

    // Function to clear error message and reset field color
    function clearError(field) {
        field.style.borderColor = '';
        var errorElement = field.nextElementSibling;
        if (errorElement && errorElement.classList.contains('error-message')) {
            errorElement.remove();
        }
    }

    // Validation functions
    function isValidUsernameOrEmail(input) {
        var regex = /^(?!\d)[a-z0-9._@]+$/;
        return regex.test(input);
    }

    function isValidPassword(password) {
        var regex = /^(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*()_+\-=\[\]{}|:;\"'<>,.?\/]).{8,}$/;
        return regex.test(password);
    }

    var allValid = true;
    var users = JSON.parse(localStorage.getItem('users')) || [];

    // Validate username
    if (!isValidUsernameOrEmail(username)) {
        showError(usernameField, 'Username must not have capital letters, must not start with a digit, and can only include ., _, @.');
        allValid = false;
    } else if (users.some(function (user) { return user.username === username; })) {
        showError(usernameField, 'Username already exists. Please choose another.');
        allValid = false;
    } else {
        clearError(usernameField);
    }

    // Validate email
    if (!isValidUsernameOrEmail(email)) {
        showError(emailField, 'Email must not have capital letters, must not start with a digit, and can only include ., _, @.');
        allValid = false;
    } else if (users.some(function (user) { return user.email === email; })) {
        showError(emailField, 'Email already exists. Please use another.');
        allValid = false;
    } else {
        clearError(emailField);
    }

    // Validate password
    if (!isValidPassword(password)) {
        showError(passwordField, 'Password must be at least 8 characters long, include 1 uppercase letter, 1 digit, and 1 special symbol.');
        allValid = false;
    } else {
        clearError(passwordField);
    }

    // If all fields are valid, proceed with signup
    if (allValid) {
        var newUser = { username: username, email: email, password: password };
        users.push(newUser);
        localStorage.setItem('users', JSON.stringify(users));

        alert('Signup successful!');
        window.location.href = 'login.html';
    }
});
