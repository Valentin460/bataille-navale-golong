package game

import "testing"

func TestNomDeLaFonction(t *testing.T) {
    // Arrange
    g := NewGame(10, []int{5})
    
    // Act
    result := g.ProcessHit(5, 5)
    
    // Assert
    if result.Result != "miss" {
        t.Errorf("Expected miss, got %s", result.Result)
    }
}