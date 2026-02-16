// Package services implements application services for post domain commands.
// my-go-backend/post_service/src/posts/application/posts/services/command_service.go
package services

import (
	"context"

	"my-go-backend/post_service/src/posts/domain"
	"my-go-backend/post_service/src/posts/domain/events"
	"my-go-backend/post_service/src/posts/domain/outbox"
	"my-go-backend/post_service/src/posts/domain/repos"
	"my-go-backend/post_service/src/shared"

	"github.com/google/uuid"
)

// PostCommandService defines the application-level interface for mutating posts.
type PostCommandService interface {
	CreatePost(ctx context.Context, title, content string, authorID, communityID uuid.UUID, isPublic bool) (*domain.PostAggregate, error)
	UpdatePost(ctx context.Context, postID uuid.UUID, newTitle, newContent string, requesterID uuid.UUID) error
	TogglePostVisibility(ctx context.Context, postID uuid.UUID, isPublic bool, requesterID uuid.UUID) error
	LikePost(ctx context.Context, postID uuid.UUID) error
	UnlikePost(ctx context.Context, postID uuid.UUID) error
	AddCommentToPost(ctx context.Context, postID uuid.UUID) error
	RemoveCommentFromPost(ctx context.Context, postID uuid.UUID) error
	DeletePost(ctx context.Context, postID uuid.UUID, requesterID uuid.UUID) error
}

var _ PostCommandService = (*PostCommandServiceImpl)(nil)

type PostCommandServiceImpl struct {
	postRepo   repos.PostCommandRepository
	outboxRepo outbox.OutboxRepository
}

func NewPostCommandService(
	postRepo repos.PostCommandRepository,
	outboxRepo outbox.OutboxRepository,
) PostCommandService {
	return &PostCommandServiceImpl{
		postRepo:   postRepo,
		outboxRepo: outboxRepo,
	}
}

// publishEvent stages a domain event in the outbox table.
// Accepts any event implementing shared.DomainEvent.
func (s *PostCommandServiceImpl) publishEvent(ctx context.Context, event shared.DomainEvent) error {
	outboxEvent, err := outbox.NewOutboxEvent(
		event.EventType(),
		event.ToMap(), // payload is the full serializable map
		uuid.MustParse(event.AggregateID()),
		event.AggregateType(),
		nil, // traceID – extend later if needed
		nil, // metadata
	)
	if err != nil {
		return shared.NewInternalServerError(err)
	}

	if err := s.outboxRepo.Save(ctx, outboxEvent); err != nil {
		return outbox.NewOutboxSaveError(
			event.EventType(),
			uuid.MustParse(event.AggregateID()),
			err.Error(),
		)
	}
	return nil
}

// ------------------ COMMAND IMPLEMENTATIONS ------------------

func (s *PostCommandServiceImpl) CreatePost(
	ctx context.Context,
	title, content string,
	authorID, communityID uuid.UUID,
	isPublic bool,
) (*domain.PostAggregate, error) {
	agg, err := domain.CreatePost(title, content, authorID, communityID, isPublic)
	if err != nil {
		return nil, err
	}

	exists, err := s.postRepo.ExistsWithTitleInCommunity(ctx, communityID, title)
	if err != nil {
		return nil, shared.NewInternalServerError(err)
	}
	if exists {
		return nil, domain.NewPostAlreadyExistsError(agg.ID().String())
	}

	if err := s.postRepo.Create(ctx, agg); err != nil {
		return nil, shared.NewInternalServerError(err)
	}

	event := events.NewPostCreatedEvent(agg)
	if err := s.publishEvent(ctx, event); err != nil {
		return nil, err
	}

	return agg, nil
}

func (s *PostCommandServiceImpl) UpdatePost(
	ctx context.Context,
	postID uuid.UUID,
	newTitle, newContent string,
	requesterID uuid.UUID,
) error {
	agg, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if err == domain.ErrPostNotFound {
			return domain.NewPostNotFoundError(postID.String())
		}
		return shared.NewInternalServerError(err)
	}

	if err := agg.UpdateContent(newTitle, newContent, requesterID); err != nil {
		return err
	}

	if err := s.postRepo.Update(ctx, agg); err != nil {
		return shared.NewInternalServerError(err)
	}

	event := events.NewPostUpdatedEvent(agg)
	if err := s.publishEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

func (s *PostCommandServiceImpl) TogglePostVisibility(
	ctx context.Context,
	postID uuid.UUID,
	isPublic bool,
	requesterID uuid.UUID,
) error {
	agg, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if err == domain.ErrPostNotFound {
			return domain.NewPostNotFoundError(postID.String())
		}
		return shared.NewInternalServerError(err)
	}

	if err := agg.ToggleVisibility(isPublic, requesterID); err != nil {
		return err
	}

	if err := s.postRepo.Update(ctx, agg); err != nil {
		return shared.NewInternalServerError(err)
	}

	event := events.NewPostVisibilityToggledEvent(agg)
	if err := s.publishEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

func (s *PostCommandServiceImpl) LikePost(ctx context.Context, postID uuid.UUID) error {
	agg, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if err == domain.ErrPostNotFound {
			return domain.NewPostNotFoundError(postID.String())
		}
		return shared.NewInternalServerError(err)
	}

	agg.AddLike()

	if err := s.postRepo.Update(ctx, agg); err != nil {
		return shared.NewInternalServerError(err)
	}

	event := events.NewPostLikedEvent(agg)
	if err := s.publishEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

func (s *PostCommandServiceImpl) UnlikePost(ctx context.Context, postID uuid.UUID) error {
	agg, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if err == domain.ErrPostNotFound {
			return domain.NewPostNotFoundError(postID.String())
		}
		return shared.NewInternalServerError(err)
	}

	if err := agg.RemoveLike(); err != nil {
		return err
	}

	if err := s.postRepo.Update(ctx, agg); err != nil {
		return shared.NewInternalServerError(err)
	}

	event := events.NewPostUnlikedEvent(agg)
	if err := s.publishEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

func (s *PostCommandServiceImpl) AddCommentToPost(ctx context.Context, postID uuid.UUID) error {
	agg, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if err == domain.ErrPostNotFound {
			return domain.NewPostNotFoundError(postID.String())
		}
		return shared.NewInternalServerError(err)
	}

	agg.AddComment()

	if err := s.postRepo.Update(ctx, agg); err != nil {
		return shared.NewInternalServerError(err)
	}

	event := events.NewPostCommentedEvent(agg)
	if err := s.publishEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

func (s *PostCommandServiceImpl) RemoveCommentFromPost(ctx context.Context, postID uuid.UUID) error {
	agg, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if err == domain.ErrPostNotFound {
			return domain.NewPostNotFoundError(postID.String())
		}
		return shared.NewInternalServerError(err)
	}

	if err := agg.RemoveComment(); err != nil {
		return err
	}

	if err := s.postRepo.Update(ctx, agg); err != nil {
		return shared.NewInternalServerError(err)
	}

	event := events.NewPostCommentRemovedEvent(agg)
	if err := s.publishEvent(ctx, event); err != nil {
		return err
	}

	return nil
}

func (s *PostCommandServiceImpl) DeletePost(ctx context.Context, postID uuid.UUID, requesterID uuid.UUID) error {
	agg, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if err == domain.ErrPostNotFound {
			return domain.NewPostNotFoundError(postID.String())
		}
		return shared.NewInternalServerError(err)
	}

	// ✅ Use == for UUID comparison (always works)
	if agg.AuthorID() != requesterID {
		return domain.NewUserNotPostAuthorError(postID.String(), requesterID.String())
	}

	event := events.NewPostDeletedEvent(agg)
	if err := s.publishEvent(ctx, event); err != nil {
		return err
	}

	if err := s.postRepo.Delete(ctx, postID); err != nil {
		if err == domain.ErrPostNotFound {
			return domain.NewPostNotFoundError(postID.String())
		}
		return shared.NewInternalServerError(err)
	}

	return nil
}
