Objective: Assess your ability to design and implement a performant, idiomatic Golang RESTful API service with PostgreSQL integration, focusing on concurrency, JSON handling, and database interaction. Challenge Description You are tasked with building a simple backend service in Go that manages a collection of "Events". Each Event has the following fields:

id (UUID, primary key)

title (string, max 100 characters)

description (string, optional)

start_time (timestamp)

end_time (timestamp)

created_at (timestamp, auto-set on creation)

Your service should expose a RESTful API with the following endpoints:
Create Event

POST /events

Accepts a JSON payload with title, description, start_time, and end_time.

Validates that title is non-empty and <= 100 characters, start_time is before end_time.

Inserts the event into a PostgreSQL database, generating a UUID for id and setting created_at to current time.

Returns the created event as JSON with HTTP 201 status.

List Events

GET /events

Returns a JSON array of all events ordered by start_time ascending.

Get Event by ID

GET /events/{id}

Returns the event with the specified UUID or 404 if not found.
