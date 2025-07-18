-- name: GetMovieByID :one
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE id = ?;

-- name: GetAllMovies :many
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
ORDER BY sort_title;

-- name: GetMoviesByQualityProfile :many
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE quality_profile_id = ?
ORDER BY sort_title;

-- name: GetMonitoredMovies :many
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE monitored = true
ORDER BY sort_title;

-- name: CreateMovie :execresult
INSERT INTO movies (
    tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    folder_name, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
    NOW(), NOW()
);

-- name: UpdateMovie :exec
UPDATE movies SET
    tmdb_id = ?,
    imdb_id = ?,
    title = ?,
    original_title = ?,
    sort_title = ?,
    year = ?,
    release_date = ?,
    runtime = ?,
    certification = ?,
    overview = ?,
    images = ?,
    genres = ?,
    studio = ?,
    path = ?,
    quality_profile_id = ?,
    monitored = ?,
    minimum_availability = ?,
    status = ?,
    has_file = ?,
    added = ?,
    folder_name = ?,
    updated_at = NOW()
WHERE id = ?;

-- name: DeleteMovie :exec
DELETE FROM movies WHERE id = ?;

-- name: CountMovies :one
SELECT COUNT(*) FROM movies;

-- name: CountMonitoredMovies :one
SELECT COUNT(*) FROM movies WHERE monitored = true;

-- name: GetMoviesWithFiles :many
SELECT
    id, tmdb_id, imdb_id, title, original_title, sort_title,
    year, release_date, runtime, certification, overview,
    images, genres, studio, path, quality_profile_id,
    monitored, minimum_availability, status, has_file,
    added, folder_name, created_at, updated_at
FROM movies
WHERE has_file = true
ORDER BY sort_title;
