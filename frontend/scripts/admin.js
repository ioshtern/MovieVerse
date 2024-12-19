const apiUrl = 'http://localhost:8080'; // Replace with your API base URL

const showResponse = (response) => {
    const responseBody = document.getElementById('response-body');
    responseBody.innerHTML = ""; // Clear previous response

    if (Array.isArray(response)) {
        // For arrays (e.g., list of movies or users)
        const headerRow = document.createElement('tr');
        if (response.length > 0) {
            const firstItem = response[0];
            for (const key in firstItem) {
                const th = document.createElement('th');
                th.textContent = key;
                headerRow.appendChild(th);
            }
            responseBody.appendChild(headerRow);

            response.forEach(item => {
                const row = document.createElement('tr');
                for (const key in item) {
                    const cell = document.createElement('td');
                    const value = Array.isArray(item[key]) ? item[key].join(", ") : item[key];
                    cell.textContent = value;
                    row.appendChild(cell);
                }
                responseBody.appendChild(row);
            });
        }
    } else {
        // For a single object (e.g., a single movie or user)
        const headerRow = document.createElement('tr');
        for (const key in response) {
            const th = document.createElement('th');
            th.textContent = key;
            headerRow.appendChild(th);
        }
        responseBody.appendChild(headerRow);

        const row = document.createElement('tr');
        for (const key in response) {
            const cell = document.createElement('td');
            const value = Array.isArray(response[key]) ? response[key].join(", ") : response[key];
            cell.textContent = value;
            row.appendChild(cell);
        }
        responseBody.appendChild(row);
    }
};

// Movie Functions
function getMovies() {
    fetch(`${apiUrl}/movies`)
        .then(res => res.json())
        .then(showResponse)
        .catch(console.error);
}

function getMovieByID(event) {
    event.preventDefault();
    const id = document.getElementById("movieID").value;
    fetch(`${apiUrl}/movies?id=${id}`)
        .then(res => res.json())
        .then(showResponse)
        .catch(console.error);
}

function createMovie(event) {
    event.preventDefault();
    const formData = {
        title: document.getElementById("title").value,
        director: document.getElementById("director").value,
        country: document.getElementById("country").value,
        genres: document.getElementById("genres").value.split(","),
        release_year: parseInt(document.getElementById("releaseYear").value),
        description: document.getElementById("description").value,
    };

    fetch(`${apiUrl}/movies`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(formData),
    }).then(res => res.json()).then(showResponse).catch(console.error);
}

function deleteMovie(event) {
    event.preventDefault();
    const id = document.getElementById("deleteMovieID").value;
    fetch(`${apiUrl}/movies?id=${id}`, {
        method: "DELETE",
    }).then(res => res.json()).then(showResponse).catch(console.error);
}

// User Functions
function getAllUsers() {
    fetch(`${apiUrl}/users`)
        .then(res => res.json())
        .then(showResponse)
        .catch(console.error);
}

function getUserByID(event) {
    event.preventDefault();
    const id = document.getElementById("userID").value;
    fetch(`${apiUrl}/users?id=${id}`)
        .then(res => res.json())
        .then(showResponse)
        .catch(console.error);
}

function deleteUser(event) {
    event.preventDefault();
    const id = document.getElementById("deleteUserID").value;
    fetch(`${apiUrl}/users?id=${id}`, {
        method: "DELETE",
    }).then(res => res.json()).then(showResponse).catch(console.error);
}