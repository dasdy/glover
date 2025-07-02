package routes_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockComponent implements the templ.Component interface for testing.
type MockComponent struct {
	RenderFunc func(ctx context.Context, w io.Writer) error
}

func (m MockComponent) Render(ctx context.Context, w io.Writer) error {
	return m.RenderFunc(ctx, w)
}

func TestSafeRenderTemplate(t *testing.T) {
	t.Run("successful render", func(t *testing.T) {
		// Create a mock component that writes "Hello, World!" to the writer
		mockComponent := MockComponent{
			RenderFunc: func(_ context.Context, w io.Writer) error {
				_, err := w.Write([]byte("Hello, World!"))
				if err != nil {
					return fmt.Errorf("failed to write data: %w", err)
				}

				return nil
			},
		}

		// Create a test response recorder
		recorder := httptest.NewRecorder()

		// Call the function
		err := routes.SafeRenderTemplate(mockComponent, recorder)

		// Assert there's no error
		require.NoError(t, err)

		// Assert the response has the correct content type
		assert.Equal(t, "text/html; charset=UTF-8", recorder.Header().Get("Content-Type"))

		// Assert the response body is correct
		assert.Equal(t, "Hello, World!", recorder.Body.String())
	})

	t.Run("render error", func(t *testing.T) {
		// Create a mock component that returns an error
		expectedErr := errors.New("render error")
		mockComponent := MockComponent{
			RenderFunc: func(_ context.Context, _ io.Writer) error {
				return expectedErr
			},
		}

		// Create a test response recorder
		recorder := httptest.NewRecorder()

		// Call the function
		err := routes.SafeRenderTemplate(mockComponent, recorder)

		// Assert the error is returned and wrapped
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not render template")

		// Assert no response was written
		assert.Empty(t, recorder.Body.String())
	})
}

func TestInitEmptyMap(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		// Setup sample data
		names := []string{"A", "B", "C"}
		locations := map[model.KeyPosition]model.Location{
			0: {RowCol: model.RowCol{Row: 0, Col: 0}},
			1: {RowCol: model.RowCol{Row: 0, Col: 1}},
			2: {RowCol: model.RowCol{Row: 1, Col: 0}},
		}

		// Call the function
		result := routes.InitEmptyMap(names, locations)

		// Assert the map has the correct entries
		assert.Len(t, result, 3)

		// Check each entry
		assert.Equal(t, 0, result[model.RowCol{Row: 0, Col: 0}].Count)
		assert.Equal(t, model.KeyPosition(0), result[model.RowCol{Row: 0, Col: 0}].Position)
		assert.Equal(t, "A", result[model.RowCol{Row: 0, Col: 0}].KeyLabel)

		assert.Equal(t, 0, result[model.RowCol{Row: 0, Col: 1}].Count)
		assert.Equal(t, model.KeyPosition(1), result[model.RowCol{Row: 0, Col: 1}].Position)
		assert.Equal(t, "B", result[model.RowCol{Row: 0, Col: 1}].KeyLabel)

		assert.Equal(t, 0, result[model.RowCol{Row: 1, Col: 0}].Count)
		assert.Equal(t, model.KeyPosition(2), result[model.RowCol{Row: 1, Col: 0}].Position)
		assert.Equal(t, "C", result[model.RowCol{Row: 1, Col: 0}].KeyLabel)
	})

	t.Run("fewer names than positions", func(t *testing.T) {
		// Setup sample data with more positions than names
		names := []string{"A", "B"}
		locations := map[model.KeyPosition]model.Location{
			0: {RowCol: model.RowCol{Row: 0, Col: 0}},
			1: {RowCol: model.RowCol{Row: 0, Col: 1}},
			2: {RowCol: model.RowCol{Row: 1, Col: 0}},
		}

		// Call the function
		result := routes.InitEmptyMap(names, locations)

		// Assert the map has entries for all positions
		assert.Len(t, result, 3)

		// Check that positions with names use those names
		assert.Equal(t, "A", result[model.RowCol{Row: 0, Col: 0}].KeyLabel)
		assert.Equal(t, "B", result[model.RowCol{Row: 0, Col: 1}].KeyLabel)

		// Check that positions without names use the default
		assert.Equal(t, "<OOB>", result[model.RowCol{Row: 1, Col: 0}].KeyLabel)
	})

	t.Run("empty inputs", func(t *testing.T) {
		// Test with empty inputs
		names := []string{}
		locations := map[model.KeyPosition]model.Location{}

		// Call the function
		result := routes.InitEmptyMap(names, locations)

		// Assert the result is an empty map
		assert.Empty(t, result)
	})
}
