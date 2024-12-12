document.addEventListener('DOMContentLoaded', function () {
    var userList = document.getElementById('userList');
    var users = JSON.parse(localStorage.getItem('users')) || [];
    var logoutButton = document.getElementById('logoutButton');

    function renderUsers() {
        userList.innerHTML = '';
        users.forEach(function (user, index) {
            var userItem = document.createElement('li');
            userItem.classList.add('list-group-item');
            userItem.innerHTML = `
                <div>
                    <strong>Username:</strong>
                    <input type="text" class="username" value="${user.username}" data-index="${index}" readonly />
                    <button class="btn btn-secondary btn-sm edit-username" data-index="${index}">Edit</button>
                </div>
                <div>
                    <strong>Email:</strong>
                    <input type="text" class="email" value="${user.email}" data-index="${index}" readonly />
                    <button class="btn btn-secondary btn-sm edit-email" data-index="${index}">Edit</button>
                </div>
                <div>
                    <strong>Password:</strong>
                    <input type="password" class="password" value="${user.password}" data-index="${index}" readonly />
                    <button class="btn btn-secondary btn-sm edit-password" data-index="${index}">Edit</button>
                    <button class="btn btn-secondary btn-sm toggle-password" data-index="${index}">Show</button>
                </div>
                <button class="btn btn-danger btn-sm delete-button" data-index="${index}">Delete</button>
            `;
            userList.appendChild(userItem);
        });

        var editUsernameButtons = document.querySelectorAll('.edit-username');
        editUsernameButtons.forEach(function (button) {
            button.addEventListener('click', handleEditUsername);
        });

        var editEmailButtons = document.querySelectorAll('.edit-email');
        editEmailButtons.forEach(function (button) {
            button.addEventListener('click', handleEditEmail);
        });

        var editPasswordButtons = document.querySelectorAll('.edit-password');
        editPasswordButtons.forEach(function (button) {
            button.addEventListener('click', handleEditPassword);
        });

        var togglePasswordButtons = document.querySelectorAll('.toggle-password');
        togglePasswordButtons.forEach(function (button) {
            button.addEventListener('click', handleTogglePassword);
        });

        var deleteButtons = document.querySelectorAll('.delete-button');
        deleteButtons.forEach(function (button) {
            button.addEventListener('click', handleDeleteUser);
        });
    }

    function handleEditUsername(event) {
        var index = event.target.getAttribute('data-index');
        var inputField = document.querySelector(`.username[data-index="${index}"]`);
        inputField.readOnly = false;
        inputField.focus();

        inputField.addEventListener('blur', function () {
            var newUsername = inputField.value.trim();
            if (newUsername) {
                users[index].username = newUsername;
                localStorage.setItem('users', JSON.stringify(users));
            }
            inputField.readOnly = true;
            renderUsers();
        });
    }

    function handleEditEmail(event) {
        var index = event.target.getAttribute('data-index');
        var inputField = document.querySelector(`.email[data-index="${index}"]`);
        inputField.readOnly = false;
        inputField.focus();

        inputField.addEventListener('blur', function () {
            var newEmail = inputField.value.trim();
            if (newEmail) {
                users[index].email = newEmail;
                localStorage.setItem('users', JSON.stringify(users));
            }
            inputField.readOnly = true;
            renderUsers();
        });
    }

    function handleEditPassword(event) {
        var index = event.target.getAttribute('data-index');
        var inputField = document.querySelector(`.password[data-index="${index}"]`);
        inputField.readOnly = false;
        inputField.focus();

        inputField.addEventListener('blur', function () {
            var newPassword = inputField.value.trim();
            if (newPassword) {
                users[index].password = newPassword;
                localStorage.setItem('users', JSON.stringify(users));
            }
            inputField.readOnly = true;
            renderUsers();
        });
    }

    function handleTogglePassword(event) {
        var index = event.target.getAttribute('data-index');
        var inputField = document.querySelector(`.password[data-index="${index}"]`);
        var type = inputField.getAttribute('type');

        if (type === 'password') {
            inputField.setAttribute('type', 'text');
            event.target.innerText = 'Hide';
        } else {
            inputField.setAttribute('type', 'password');
            event.target.innerText = 'Show';
        }
    }

    function handleDeleteUser(event) {
        var index = event.target.getAttribute('data-index');
        users.splice(index, 1);
        localStorage.setItem('users', JSON.stringify(users));
        renderUsers();
    }

    logoutButton.addEventListener('click', function () {
        localStorage.removeItem('loggedInUser');
        window.location.href = 'login.html'; 
    });

    renderUsers();
});
