package ports

const (
	QueryInsertPost = `
		INSERT INTO posts (
			id, author_id, community_id, title, content, is_public,
			likes_count, comment_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	QueryUpdatePost = `
		UPDATE posts SET
			title = $1, content = $2, is_public = $3,
			likes_count = $4, comment_count = $5, updated_at = $6
		WHERE id = $7
	`

	QueryDeletePost = `
		DELETE FROM posts WHERE id = $1
	`

	QueryExistsPostWithTitleInCommunity = `
		SELECT COUNT(*) FROM posts
		WHERE community_id = $1 AND LOWER(title) = LOWER($2)
	`
	QuerySelectPostByID = `
		SELECT id, author_id, community_id, title, content, is_public,
		       likes_count, comment_count, created_at, updated_at
		FROM posts WHERE id = $1 LIMIT 1
	`

	QuerySelectPostsByAuthor = `
		SELECT id, author_id, community_id, title, content, is_public,
		       likes_count, comment_count, created_at, updated_at
		FROM posts
		WHERE author_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	QueryCountPostsByAuthor = `
		SELECT COUNT(*) FROM posts WHERE author_id = $1
	`

	QuerySelectPostsByCommunity = `
		SELECT id, author_id, community_id, title, content, is_public,
			likes_count, comment_count, created_at, updated_at
		FROM posts
		WHERE community_id = $1
		AND (is_public = true OR ($2::uuid IS NOT NULL AND author_id = $2::uuid))
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	QuerySearchPosts = `
		SELECT id, author_id, community_id, title, content, is_public,
		       likes_count, comment_count, created_at, updated_at
		FROM posts
		WHERE to_tsvector('english', title || ' ' || content) @@ plainto_tsquery('english', $1)
		ORDER BY ts_rank(to_tsvector('english', title || ' ' || content), plainto_tsquery('english', $1)) DESC
		LIMIT $2 OFFSET $3
	`

	QueryExistsPostByID = `
		SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)
	`
)
