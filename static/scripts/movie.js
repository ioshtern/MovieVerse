document.addEventListener('DOMContentLoaded', function () {
    var loginButton = document.getElementById('loginButton');
    var signupButton = document.getElementById('signupButton');
    var submitButton = document.querySelector('button.btn-primary');
    var reviewTextArea = document.querySelector('textarea');
    var reviewList = document.querySelector('.list-group');
    var averageRatingLabel = document.querySelector('h4');
    var stars = document.querySelectorAll('.fa-star');
    var movieId = window.movieId;
    var loggedInUser = JSON.parse(localStorage.getItem('loggedInUser'));

    if (loggedInUser) {
        loginButton.style.display = 'none';
        signupButton.style.display = 'none';
    }

    var totalRating = 0;
    var reviewCount = 0;

    var reviews = JSON.parse(localStorage.getItem('reviews')) || {};
    var movieReviews = reviews[movieId] || [];

    movieReviews.forEach(function (review) {
        var reviewItem = document.createElement('li');
        reviewItem.classList.add('list-group-item');
        reviewItem.innerHTML = '<strong>' + review.username + ':</strong><em> ' + review.rating + ' stars</em><p>' + review.text + '</p>';
        reviewList.appendChild(reviewItem);
        totalRating += review.rating;
    });

    if (movieReviews.length > 0) {
        reviewCount = movieReviews.length;
        var averageRating = (totalRating / reviewCount).toFixed(1);
        averageRatingLabel.innerText = 'Average Rating: ' + averageRating;
    } else {
        averageRatingLabel.innerText = 'Average Rating: 0.0';
    }

    var currentRating = 0;

    stars.forEach(function (star) {
        star.addEventListener('click', function () {
            currentRating = parseInt(this.getAttribute('data-value'));
            updateStars(currentRating);
        });
    });

    function updateStars(rating) {
        stars.forEach(function (star) {
            star.style.color = star.getAttribute('data-value') <= rating ? 'gold' : 'gray';
        });
    }

    submitButton.addEventListener('click', function () {
        if (!loggedInUser) {
            alert('Please log in to submit a review.');
            return;
        }

        var reviewText = reviewTextArea.value.trim();

        if (currentRating > 0 && reviewText !== '') {
            var review = {
                username: loggedInUser.username,
                rating: currentRating,
                text: reviewText,
                movieId: movieId
            };

            var reviewItem = document.createElement('li');
            reviewItem.classList.add('list-group-item');
            reviewItem.innerHTML = '<strong>' + review.username + ' - Rating: ' + review.rating + ' stars</strong><p>' + review.text + '</p>';
            reviewList.appendChild(reviewItem);

            totalRating += currentRating;
            reviewCount++;

            var averageRating = (totalRating / reviewCount).toFixed(1);
            averageRatingLabel.innerText = 'Average Rating: ' + averageRating;

            if (!reviews[movieId]) {
                reviews[movieId] = [];
            }
            reviews[movieId].push(review);
            localStorage.setItem('reviews', JSON.stringify(reviews));

            reviewTextArea.value = '';
            updateStars(0);
            currentRating = 0;
        } else {
            alert('Please select a rating and write a review.');
        }
    });
});

var body = document.querySelector('body');

var container = document.createElement('div');
container.classList.add('container', 'bg-white', 'mt-1', 'pt-2');

var row = document.createElement('div');
row.classList.add('row', 'border', 'border-white', 'border-5', 'ps-5');
var col = document.createElement('div');
col.classList.add('col-lg-11', 'align-middle');

var title = document.createElement('h3');
title.innerText = 'Rate and Review';
col.appendChild(title);

var ratingContainer = document.createElement('div');
ratingContainer.style.display = 'flex';
ratingContainer.style.cursor = 'pointer';

var starsArray = [];
for (var i = 1; i <= 5; i++) {
    var star = document.createElement('i');
    star.classList.add('fa', 'fa-star');
    star.style.fontSize = '2rem';
    star.style.color = 'gray';
    star.setAttribute('data-value', i);

    star.addEventListener('click', function () {
        currentRating = this.getAttribute('data-value');
        updateStars(currentRating);
    });

    ratingContainer.appendChild(star);
    starsArray.push(star);
}

function updateStars(rating) {
    starsArray.forEach(function (star) {
        star.style.color = star.getAttribute('data-value') <= rating ? 'gold' : 'gray';
    });
}

col.appendChild(ratingContainer);

var averageRatingContainer = document.createElement('div');
averageRatingContainer.style.marginTop = '10px';

var averageRatingLabel = document.createElement('h4');
averageRatingLabel.innerText = 'Average Rating: 0.0';
averageRatingContainer.appendChild(averageRatingLabel);

col.appendChild(averageRatingContainer);

var reviewTextArea = document.createElement('textarea');
reviewTextArea.placeholder = 'Write your review...';
reviewTextArea.classList.add('form-control', 'mb-2');
reviewTextArea.style.marginTop = '10px';
reviewTextArea.style.marginBottom = '10px';
reviewTextArea.style.padding = '10px';
reviewTextArea.style.border = '1px solid #ccc';
reviewTextArea.style.borderRadius = '5px';

col.appendChild(reviewTextArea);

var submitButton = document.createElement('button');
submitButton.innerText = 'Submit Review';
submitButton.classList.add('btn', 'btn-primary');
submitButton.style.cursor = 'pointer';

col.appendChild(submitButton);

var reviewSection = document.createElement('div');
reviewSection.style.marginTop = '20px';

var reviewTitle = document.createElement('h4');
reviewTitle.innerText = 'Reviews';
reviewSection.appendChild(reviewTitle);

var reviewList = document.createElement('ul');
reviewList.classList.add('list-group');
reviewSection.appendChild(reviewList);

col.appendChild(reviewSection);

row.appendChild(col);
container.appendChild(row);
body.appendChild(container);
