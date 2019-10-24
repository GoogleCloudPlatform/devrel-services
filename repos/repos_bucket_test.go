package repos

import (
	"testing"
	"context"
)

func TestGetRepos(t *testing.T){
	 
	expectedRepositories := []TrackedRepository{TrackedRepository{}}
	underTest := NewBucketRepo("Pokedex","jantho")
	ctx := context.WithValue(context.Background(), "Pikachu","Go")
	repos ,err := underTest.getRepos(ctx)

	if err != nil {
		t.Error("Impossible to get TrackedRepository")
	}

	 if len(repos) != len(expectedRepositories) {
		t.Error("The repository length doesn't correspond.")
	}

}	
