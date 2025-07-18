-- name: GetQualityProfileByID :one
SELECT id, name, cutoff, items, created_at, updated_at
FROM quality_profiles
WHERE id = ?;

-- name: GetAllQualityProfiles :many
SELECT id, name, cutoff, items, created_at, updated_at
FROM quality_profiles
ORDER BY name;

-- name: CreateQualityProfile :execresult
INSERT INTO quality_profiles (name, cutoff, items, created_at, updated_at)
VALUES (?, ?, ?, NOW(), NOW());

-- name: UpdateQualityProfile :exec
UPDATE quality_profiles SET
    name = ?,
    cutoff = ?,
    items = ?,
    updated_at = NOW()
WHERE id = ?;

-- name: DeleteQualityProfile :exec
DELETE FROM quality_profiles WHERE id = ?;

-- name: CountQualityProfiles :one
SELECT COUNT(*) FROM quality_profiles;
