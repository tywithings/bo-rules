package rules

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoyaleRulesetInterface(t *testing.T) {
	var _ Ruleset = (*RoyaleRuleset)(nil)
}

func TestRoyaleDefaultSanity(t *testing.T) {
	boardState := &BoardState{
		Snakes: []Snake{
			{ID: "1", Body: []Point{{X: 0, Y: 0}}},
			{ID: "2", Body: []Point{{X: 0, Y: 1}}},
		},
	}
	r := RoyaleRuleset{StandardRuleset: StandardRuleset{HazardDamagePerTurn: 1}, ShrinkEveryNTurns: 0}
	_, err := r.CreateNextBoardState(boardState, []SnakeMove{{"1", "right"}, {"2", "right"}})
	require.Error(t, err)
	require.Equal(t, errors.New("royale game can't shrink more frequently than every turn"), err)

	r = RoyaleRuleset{ShrinkEveryNTurns: 1}
	_, err = r.CreateNextBoardState(boardState, []SnakeMove{})
	require.Error(t, err)
	require.Equal(t, errors.New("royale damage per turn must be greater than zero"), err)

	r = RoyaleRuleset{StandardRuleset: StandardRuleset{HazardDamagePerTurn: 1}, ShrinkEveryNTurns: 1}
	boardState, err = r.CreateNextBoardState(boardState, []SnakeMove{})
	require.NoError(t, err)
	require.Len(t, boardState.Hazards, 0)
}

func TestRoyaleName(t *testing.T) {
	r := RoyaleRuleset{}
	require.Equal(t, "royale", r.Name())
}

func TestRoyaleHazards(t *testing.T) {
	seed := int64(25543234525)
	tests := []struct {
		Width             int
		Height            int
		Turn              int
		ShrinkEveryNTurns int
		Error             error
		ExpectedHazards   []Point
	}{
		{Error: errors.New("royale game can't shrink more frequently than every turn")},
		{ShrinkEveryNTurns: 1, ExpectedHazards: []Point{}},
		{Turn: 1, ShrinkEveryNTurns: 1, ExpectedHazards: []Point{}},
		{Width: 3, Height: 3, Turn: 1, ShrinkEveryNTurns: 10, ExpectedHazards: []Point{}},
		{Width: 3, Height: 3, Turn: 9, ShrinkEveryNTurns: 10, ExpectedHazards: []Point{}},
		{
			Width: 3, Height: 3, Turn: 10, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 11, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 19, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 20, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 2}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 31, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 1}, {1, 2}, {2, 1}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 42, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 53, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 64, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 6987, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		},
	}

	for _, test := range tests {
		b := &BoardState{
			Turn:   test.Turn - 1,
			Width:  test.Width,
			Height: test.Height,
		}
		settings := Settings{
			HazardDamagePerTurn: 1,
			RoyaleSettings: RoyaleSettings{
				ShrinkEveryNTurns: test.ShrinkEveryNTurns,
			},
		}.WithSeed(seed)

		_, err := PopulateHazardsRoyale(b, settings, mockSnakeMoves())
		require.Equal(t, test.Error, err)
		if err == nil {
			// Obstacles should match
			require.Equal(t, test.ExpectedHazards, b.Hazards)
			for _, expectedP := range test.ExpectedHazards {
				wasFound := false
				for _, actualP := range b.Hazards {
					if expectedP == actualP {
						wasFound = true
						break
					}
				}
				require.True(t, wasFound)
			}
		}
	}
}

// Checks that hazards get placed
// also that:
// - snakes move properly
// - snake gets health from eating
// - food gets consumed
// - health is decreased
var royaleCaseHazardsPlaced = gameTestCase{
	"Royale Case Hazards Placed",
	&BoardState{
		Width:  10,
		Height: 10,
		Snakes: []Snake{
			{
				ID:     "one",
				Body:   []Point{{1, 1}, {1, 2}},
				Health: 100,
			},
			{
				ID:     "two",
				Body:   []Point{{3, 4}, {3, 3}},
				Health: 100,
			},
			{
				ID:              "three",
				Body:            []Point{},
				Health:          100,
				EliminatedCause: EliminatedByOutOfBounds,
			},
		},
		Food:    []Point{{0, 0}, {1, 0}},
		Hazards: []Point{},
	},
	[]SnakeMove{
		{ID: "one", Move: MoveDown},
		{ID: "two", Move: MoveUp},
		{ID: "three", Move: MoveLeft}, // Should be ignored
	},
	nil,
	&BoardState{
		Width:  10,
		Height: 10,
		Snakes: []Snake{
			{
				ID:     "one",
				Body:   []Point{{1, 0}, {1, 1}, {1, 1}},
				Health: 100,
			},
			{
				ID:     "two",
				Body:   []Point{{3, 5}, {3, 4}},
				Health: 99,
			},
			{
				ID:              "three",
				Body:            []Point{},
				Health:          100,
				EliminatedCause: EliminatedByOutOfBounds,
			},
		},
		Food:    []Point{{0, 0}},
		Hazards: []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}, {X: 4, Y: 0}, {X: 5, Y: 0}, {X: 6, Y: 0}, {X: 7, Y: 0}, {X: 8, Y: 0}, {X: 9, Y: 0}},
	},
}

func TestRoyaleCreateNextBoardState(t *testing.T) {
	// add expected hazards to the standard cases that need them
	s1 := standardCaseMoveEatAndGrow.clone()
	s1.expectedState.Hazards = []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}, {X: 4, Y: 0}, {X: 5, Y: 0}, {X: 6, Y: 0}, {X: 7, Y: 0}, {X: 8, Y: 0}, {X: 9, Y: 0}}
	s2 := standardMoveAndCollideMAD.clone()
	s2.expectedState.Hazards = []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}, {X: 4, Y: 0}, {X: 5, Y: 0}, {X: 6, Y: 0}, {X: 7, Y: 0}, {X: 8, Y: 0}, {X: 9, Y: 0}}

	cases := []gameTestCase{
		// inherits these test cases from standard
		standardCaseErrNoMoveFound,
		standardCaseErrZeroLengthSnake,
		*s1,
		*s2,
		royaleCaseHazardsPlaced,
	}
	r := RoyaleRuleset{
		StandardRuleset: StandardRuleset{
			HazardDamagePerTurn: 1,
		},
		ShrinkEveryNTurns: 1,
	}
	rb := NewRulesetBuilder().WithParams(map[string]string{
		ParamGameType:            GameTypeRoyale,
		ParamHazardDamagePerTurn: "1",
		ParamShrinkEveryNTurns:   "1",
	}).WithSeed(1234)
	for _, gc := range cases {
		rand.Seed(1234)
		gc.requireValidNextState(t, &r)
		// also test a RulesBuilder constructed instance
		gc.requireValidNextState(t, rb.Ruleset())
		// also test a pipeline with the same settings
		gc.requireValidNextState(t, rb.PipelineRuleset(GameTypeRoyale, NewPipeline(royaleRulesetStages...)))
	}
}
