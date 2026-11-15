//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"

	"main/internal/config"
	"main/internal/infrastructure/database"
	"main/internal/modules/announcement/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)