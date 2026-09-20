package quiz

import (
	"errors"
	"testing"

	"github.com/tishka/tishka/internal/models"
)

func sampleTest() *models.Test {
	return &models.Test{
		Slug: "sample",
		Questions: []models.Question{
			{
				ID: "q1",
				Options: []models.Option{
					{ID: "a", Tags: []string{"cozy"}},
					{ID: "b", Tags: []string{"bold"}},
				},
			},
			{
				ID: "q2",
				Options: []models.Option{
					{ID: "a", Tags: []string{"cozy"}},
					{ID: "b", Tags: []string{"bold"}},
				},
			},
		},
		Results: []models.Result{
			{ID: "r-cozy", Tag: "cozy"},
			{ID: "r-bold", Tag: "bold"},
		},
	}
}

func TestCompute_PicksHighestScoringTag(t *testing.T) {
	e := NewEngine()
	test := sampleTest()

	res, err := e.Compute(test, Answers{"q1": "a", "q2": "a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != "r-cozy" {
		t.Errorf("got result %q, want r-cozy", res.ID)
	}
}

func TestCompute_TieBreaksToFirstDeclaredResult(t *testing.T) {
	e := NewEngine()
	test := sampleTest()

	// один голос за cozy, один за bold -> ничья, должен победить r-cozy,
	// т.к. он объявлен первым в test.Results.
	res, err := e.Compute(test, Answers{"q1": "a", "q2": "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != "r-cozy" {
		t.Errorf("got result %q, want r-cozy (first declared on tie)", res.ID)
	}
}

func TestCompute_IncompleteAnswers(t *testing.T) {
	e := NewEngine()
	test := sampleTest()

	_, err := e.Compute(test, Answers{"q1": "a"})
	if !errors.Is(err, ErrIncompleteAnswers) {
		t.Errorf("got err %v, want ErrIncompleteAnswers", err)
	}
}

func TestCompute_UnknownOption(t *testing.T) {
	e := NewEngine()
	test := sampleTest()

	_, err := e.Compute(test, Answers{"q1": "zzz", "q2": "a"})
	if !errors.Is(err, ErrUnknownOption) {
		t.Errorf("got err %v, want ErrUnknownOption", err)
	}
}

func TestModelsValidate_CatchesMissingResultForTag(t *testing.T) {
	test := sampleTest()
	test.Results = []models.Result{{ID: "r-cozy", Tag: "cozy"}} // "bold" остался без результата

	if err := test.Validate(); err == nil {
		t.Error("ожидали ошибку валидации из-за непокрытого тега bold, получили nil")
	}
}
