# Go-Testing Rune

## Purpose

Go testing patterns for Gentleman.Dots, including Bubbletea TUI testing. Use when writing Go tests, using teatest, or adding test coverage.

## When to Use

Use this rune when:
- Writing Go unit tests
- Testing Bubbletea TUI components
- Creating table-driven tests
- Adding integration tests
- Using golden file testing

## Critical Patterns

### Pattern 1: Table-Driven Tests

Standard Go pattern for multiple test cases:

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "hello",
            expected: "HELLO",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ProcessInput(tt.input)

            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

### Pattern 2: Bubbletea Model Testing

Test Model state transitions directly:

```go
func TestModelUpdate(t *testing.T) {
    m := NewModel()

    // Simulate key press
    newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
    m = newModel.(Model)

    if m.Screen != ScreenMainMenu {
        t.Errorf("expected ScreenMainMenu, got %v", m.Screen)
    }
}
```

### Pattern 3: Teatest Integration Tests

Use Charmbracelet's teatest for TUI testing:

```go
func TestInteractiveFlow(t *testing.T) {
    m := NewModel()
    tm := teatest.NewTestModel(t, m)

    // Send keys
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

    // Wait for model to update
    tm.WaitFinished(t, teatest.WithDuration(time.Second))

    // Get final model
    finalModel := tm.FinalModel(t).(Model)

    if finalModel.Screen != ExpectedScreen {
        t.Errorf("wrong screen: got %v", finalModel.Screen)
    }
}
```

### Pattern 4: Golden File Testing

Compare output against saved "golden" files:

```go
func TestOSSelectGolden(t *testing.T) {
    m := NewModel()
    m.Screen = ScreenOSSelect
    m.Width = 80
    m.Height = 24

    output := m.View()

    golden := filepath.Join("testdata", "TestOSSelectGolden.golden")

    if *update {
        os.WriteFile(golden, []byte(output), 0644)
    }

    expected, _ := os.ReadFile(golden)
    if output != string(expected) {
        t.Errorf("output doesn't match golden file")
    }
}
```

## Decision Tree

```
Testing a function?
├── Pure function? → Table-driven test
├── Has side effects? → Mock dependencies
├── Returns error? → Test both success and error cases
└── Complex logic? → Break into smaller testable units

Testing TUI component?
├── State change? → Test Model.Update() directly
├── Full flow? → Use teatest.NewTestModel()
├── Visual output? → Use golden file testing
└── Key handling? → Send tea.KeyMsg

Testing system/exec?
├── Mock os/exec? → Use interface + mock
├── Real commands? → Integration test with --short skip
└── File operations? → Use t.TempDir()
```

## Test File Organization

```
installer/internal/tui/
├── model.go
├── model_test.go           # Model tests
├── update.go
├── update_test.go          # Update handler tests
├── view.go
├── view_test.go            # View rendering tests
├── teatest_test.go         # Teatest integration tests
├── comprehensive_test.go   # Full flow tests
├── testdata/
│   ├── TestOSSelectGolden.golden
│   └── TestViewGolden.golden
└── trainer/
    ├── types.go
    ├── types_test.go
    ├── exercises.go
    ├── exercises_test.go
    └── simulator_test.go
```

## Commands

```bash
go test ./...                           # Run all tests
go test -v ./internal/tui/...          # Verbose TUI tests
go test -run TestNavigation             # Run specific test
go test -cover ./...                    # With coverage
go test -update ./...                   # Update golden files
go test -short ./...                    # Skip integration tests
```

## Resources

- **TUI Tests**: See `installer/internal/tui/*_test.go`
- **Trainer Tests**: See `installer/internal/tui/trainer/*_test.go`
- **System Tests**: See `installer/internal/system/*_test.go`
- **Golden Files**: See `installer/internal/tui/testdata/`
- **Teatest Docs**: https://github.com/charmbracelet/bubbletea/tree/master/teatest
