// my-go-backend/post_service/src/posts/application/posts/handlers/query_handlers.go
package handlers

import (
	"context"
	"my-go-backend/post_service/external"
	services "my-go-backend/post_service/external/services" // ← now imports the package with interface
	postappservices "my-go-backend/post_service/src/posts/application/posts/services"

	"github.com/google/uuid"
)

// PostQueryHandler uses interfaces for testability
type PostQueryHandler struct {
	postQueries postappservices.PostQueryService
	userClient  services.UserQueryService // ← INTERFACE, not *UserAPIClient
}

func NewPostQueryHandler(
	postQueries postappservices.PostQueryService,
	userClient services.UserQueryService, // ← INTERFACE
) *PostQueryHandler {
	return &PostQueryHandler{
		postQueries: postQueries,
		userClient:  userClient,
	}
}

// GetPostWithAuthor retrieves a single post by ID and enriches it with author details.
func (h *PostQueryHandler) GetPostWithAuthor(ctx context.Context, postID uuid.UUID) (*PostResponseDTO, error) {
	postView, err := h.postQueries.GetPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	var authorView external.UserView
	author, err := h.userClient.GetUserByID(postView.AuthorID)
	if err != nil {
		authorView = external.UserView{UserID: postView.AuthorID}
	} else {
		authorView = *author
	}

	dto := ToPostResponseDTO(postView, authorView)
	return &dto, nil
}

// GetPostsByAuthorWithAuthors retrieves posts by an author and enriches each with author data.
func (h *PostQueryHandler) GetPostsByAuthorWithAuthors(
	ctx context.Context,
	authorID uuid.UUID,
	limit, offset int,
) (*PostListResponseDTO, error) {
	posts, err := h.postQueries.GetPostsByAuthor(ctx, authorID, limit, offset)
	if err != nil {
		return nil, err
	}

	totalCount, err := h.postQueries.GetPostCountByAuthor(ctx, authorID)
	if err != nil {
		return nil, err
	}

	var authorView external.UserView
	author, err := h.userClient.GetUserByID(authorID)
	if err != nil {
		authorView = external.UserView{UserID: authorID}
	} else {
		authorView = *author
	}

	authorsMap := make(map[uuid.UUID]external.UserView, len(posts))
	for _, p := range posts {
		authorsMap[p.AuthorID] = authorView
	}

	page := offset/limit + 1
	if offset%limit != 0 {
		page++
	}

	dto := ToPostListResponseDTO(posts, authorsMap, page, limit, totalCount)
	return &dto, nil
}

// GetPostsByCommunityWithAuthors retrieves posts in a community and enriches each with its author.
func (h *PostQueryHandler) GetPostsByCommunityWithAuthors(
	ctx context.Context,
	communityID uuid.UUID,
	requesterID *uuid.UUID,
	limit, offset int,
) (*PostListResponseDTO, error) {
	posts, err := h.postQueries.GetPostsByCommunity(ctx, communityID, requesterID, limit, offset)
	if err != nil {
		return nil, err
	}

	authorIDs := make(map[uuid.UUID]struct{})
	for _, p := range posts {
		authorIDs[p.AuthorID] = struct{}{}
	}

	authorsMap := make(map[uuid.UUID]external.UserView)
	for id := range authorIDs {
		author, err := h.userClient.GetUserByID(id)
		if err != nil {
			authorsMap[id] = external.UserView{UserID: id}
		} else {
			authorsMap[id] = *author
		}
	}

	// ⚠️ Temporary total count approximation
	totalCount := offset + len(posts)

	page := offset/limit + 1
	if offset%limit != 0 {
		page++
	}

	dto := ToPostListResponseDTO(posts, authorsMap, page, limit, totalCount)
	return &dto, nil
}

// SearchPostsEnriched performs a search and enriches results with author data.
func (h *PostQueryHandler) SearchPostsEnriched(
	ctx context.Context,
	query string,
	limit, offset int,
) (*PostListResponseDTO, error) {
	if query == "" {
		return &PostListResponseDTO{
			Posts:      []PostResponseDTO{},
			TotalCount: 0,
			Page:       1,
			PageSize:   limit,
			HasMore:    false,
		}, nil
	}

	posts, err := h.postQueries.SearchPosts(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	authorIDs := make(map[uuid.UUID]struct{})
	for _, p := range posts {
		authorIDs[p.AuthorID] = struct{}{}
	}

	authorsMap := make(map[uuid.UUID]external.UserView)
	for id := range authorIDs {
		author, err := h.userClient.GetUserByID(id)
		if err != nil {
			authorsMap[id] = external.UserView{UserID: id}
		} else {
			authorsMap[id] = *author
		}
	}

	totalCount := offset + len(posts)
	page := offset/limit + 1
	if offset%limit != 0 {
		page++
	}

	dto := ToPostListResponseDTO(posts, authorsMap, page, limit, totalCount)
	return &dto, nil
}
