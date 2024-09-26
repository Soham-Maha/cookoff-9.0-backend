-- name: CreateSubmissionStatus :exec
INSERT INTO submission_results (id, submission_id, runtime, memory, description)
VALUES ($1, $2, $3, $4, $5);