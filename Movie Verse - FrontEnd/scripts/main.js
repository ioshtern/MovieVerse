document.addEventListener('DOMContentLoaded', function() {
    function updateAuthButtons() {
        const authButtonsContainer = document.querySelector('.navbar-collapse');
        const isLoggedIn = localStorage.getItem('loggedInUser');

        if (isLoggedIn) {
            authButtonsContainer.innerHTML = `
                <form class="me-3 mx-auto" role="search">
                    <input class="form-control me-2" type="search" placeholder="Search..." aria-label="Search">
                </form>
                <button class="btn btn-outline-light" type="submit">Search</button>
                <button id="logoutButton" class="btn btn-outline-light ms-2">Logout</button>
            `;

            const logoutButton = document.getElementById('logoutButton');
            logoutButton.addEventListener('click', function() {
                localStorage.removeItem('loggedInUser'); 
                location.reload(); 
            });
        }
    }
    updateAuthButtons();
});



const movieCards = document.querySelectorAll('.card.bg-secondary.p-1.movie');


function handleMouseEnter() {
    this.style.transform = 'scale(1.03)'; 
    this.style.transition = 'transform 0.3s ease'; 
}

function handleMouseLeave() {
    this.style.transform = 'scale(1)'; 
}

movieCards.forEach(function(card) {
    card.addEventListener('mouseenter', handleMouseEnter);
    card.addEventListener('mouseleave', handleMouseLeave);
});
