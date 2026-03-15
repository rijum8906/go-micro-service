package request

import "mime/multipart"

type UpdateProfileRequest struct {
	ProfileID   string   `form:"profileId" binding:"required,uuid4"`
	FirstName   *string  `form:"firstName" binding:"omitempty,min=1,max=20"`
	LastName    *string  `form:"lastName"  binding:"omitempty,min=1,max=20"`
	DisplayName *string  `form:"displayName"  binding:"omitempty,min=1,max=20"`
	AvatarURL   *string  `binding:"omitempty,max=255"`
	Metadata    Metadata `form:"metadata"  binding:"required"`

	Avatar *multipart.FileHeader `form:"avatar"`
}

type GetProfileRequest struct {
	ProfileID string   `form:"profileId" binding:"required,uuid4"`
	Metadata  Metadata `form:"metadata"  binding:"required"`
}

type DeleteProfileRequest struct {
	Metadata Metadata `form:"metadata"  binding:"required"`
}

type CreateProfileRequest struct {
	FirstName   string   `form:"firstName" binding:"required,min=1,max=20"`
	LastName    string   `form:"lastName"  binding:"required,min=1,max=20"`
	DisplayName *string  `form:"displayName"  binding:"omitempty,min=1,max=20"`
	AvatarURL   *string  `form:"avatarUrl"  binding:"omitempty,max=255"`
	Metadata    Metadata `form:"metadata"  binding:"required"`

	Avatar *multipart.FileHeader `form:"avatar"`
}
