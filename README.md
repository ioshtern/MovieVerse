# MovieVerse
## Project Description
MovieVerse is a dynamic web application designed to provide users with a platform to explore, review, and interact with movies. The project allows users to view movie details, submit reviews, and manage user interactions seamlessly. It is targeted at movie enthusiasts who want a centralized place to discover and discuss films.

## Team Members
- Bekzat Nabiev 
- Davlat Sunnatov
- Alan Yersainov

## Screenshot
![image](https://github.com/user-attachments/assets/f2ac9c0f-2842-4813-a888-f7e8fb764114)

## How to Start the Project
Follow the steps below to set up and run the project:

1. Clone the Repository
   bash
   git clone https://github.com/aituforever/MovieVerse
   cd MovieVerse
   
2. Set Up the Backend
   - Ensure you have Golang and PostgreSQL installed.
   - Configure the database connection.
     - Create a PostgreSQL database using sql script
     - Update the connection settings (e.g., port number, by default is 5432) in your Golang files.
     
3. Run the Backend Server
   bash
   go run main.go
   
   This will start the backend server on your desired port (e.g., localhost:8080).

4. Set Up the Frontend
   - Open the HTML file (admin.html) in any browser.
   - The webpage will connect to the backend server to fetch and display data.

5. Testing the Website
   - Use the browser to interact with the MovieVerse app.
   - Test movie viewing, review submissions, and user interactions.

6. Testing with Postman
   - Open Postman and set up the following requests to test the backend API:
     - GET: Retrieve all movies
              GET http://localhost:8080/movies
       
     - POST: Add a new movie
              POST http://localhost:8080/movies
       Body (raw):
       {
        "id": 3,
         "title": "The Dark Knight",
         "director": "Christopher Nolan",
         "country": "USA",
         "genres": [ "Action", "Crime", "Drama" ],
         "release_year": 2008,
         "description": "Batman faces the Joker, a criminal mastermind causing chaos in Gotham City."
       };
       
     - PUT: Update a movie
              PUT http://localhost:8080/movies?id={id}
       Body (JSON):
       {
         "title": "The Dark Knight 2",
         "director": "Christopher Nolan",
         "country": "Canada",
         "genres": [ "Action", "Crime", "Drama" ],
         "release_year": 2025,
         "description": "Batman faces the Joker, a criminal mastermind causing chaos in Gotham City again."
       };
       
     - DELETE: Delete a movie
              DELETE http://localhost:8080/movies?id={id}
       

   - Verify the responses for each request to ensure the API is functioning correctly.

## Tools and Resources
- Golang: Backend server development
- PostgreSQL: Database for storing movie data and reviews
- HTML, CSS, JavaScript: Frontend design and functionality
- Postman: API testing tool
- GitHub: Version control and project collaboration

---
Enjoy exploring movies with MovieVerse!
