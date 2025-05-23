package components_test

import (
	"testing"

	"github.com/dasdy/glover/model"
	"github.com/dasdy/glover/web/components"
	"github.com/stretchr/testify/assert"
)

func TestLocation_ToTransform(t *testing.T) {
	tests := []struct {
		name     string
		location model.Location
		want     string
	}{
		{
			name:     "zero translation",
			location: model.Location{X: 0, Y: 0, R: 0},
			want:     "translate(0.00, 0.00)",
		},
		{
			name:     "positive translation",
			location: model.Location{X: 1, Y: 2, R: 0},
			want:     "translate(80.00, 160.00)",
		},
		{
			name:     "translation with rotation",
			location: model.Location{X: 1, Y: 2, R: 45},
			want:     "translate(80.00, 160.00) rotate(45.00 0.00 0.00)",
		},
		{
			name:     "negative values",
			location: model.Location{X: -1.5, Y: -2.5, R: -90},
			want:     "translate(-120.00, -200.00) rotate(-90.00 0.00 0.00)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := components.ToTransform(&tt.location); got != tt.want {
				t.Errorf("Location.ToTransform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocation_ToTransformOrigin(t *testing.T) {
	tests := []struct {
		name     string
		location model.Location
		want     []float64
	}{
		{
			name:     "zero rotation origin",
			location: model.Location{Rx: 0, Ry: 0},
			want:     []float64{0, 0},
		},
		{
			name:     "positive rotation origin",
			location: model.Location{Rx: 1, Ry: 2},
			want:     []float64{1, 2},
		},
		{
			name:     "negative rotation origin",
			location: model.Location{Rx: -1.5, Ry: -2.5},
			want:     []float64{-1.5, -2.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := components.ToTransformOrigin(&tt.location)

			assert.Equal(t, tt.want, []float64{gotX, gotY})
		})
	}
}

func TestRenderContext_ViewBoxSize(t *testing.T) {
	tests := []struct {
		name     string
		context  components.RenderContext
		wantSize string
	}{
		{
			name: "empty grid",
			context: components.RenderContext{
				TotalCols: 0,
				TotalRows: 0,
				Items:     []components.Item{},
			},
			wantSize: "0 0 0 240",
		},
		{
			name: "regular grid without items",
			context: components.RenderContext{
				TotalCols: 3,
				TotalRows: 2,
				Items:     []components.Item{},
			},
			wantSize: "0 0 240 400",
		},
		{
			name: "grid with items within bounds",
			context: components.RenderContext{
				TotalCols: 3,
				TotalRows: 2,
				Items: []components.Item{
					{Location: model.Location{X: 1, Y: 1}},
					{Location: model.Location{X: 2, Y: 1}},
				},
			},
			wantSize: "0 0 240 400",
		},
		{
			name: "grid with items exceeding bounds",
			context: components.RenderContext{
				TotalCols: 3,
				TotalRows: 2,
				Items: []components.Item{
					{Location: model.Location{X: 4, Y: 3}},
					{Location: model.Location{X: 2, Y: 1}},
				},
			},
			wantSize: "0 0 320 480",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.context.ViewBoxSize(); got != tt.wantSize {
				t.Errorf("components.RenderContext.ViewBoxSize() = %v, want %v", got, tt.wantSize)
			}
		})
	}
}
