package layer0

import (
	"testing"
	"time"
)

func TestNewWork(t *testing.T) {
	id := WorkID("test-work")
	workType := WorkTypeTask
	name := "Test Work"

	work := NewWork(id, workType, name)

	if work.GetID() != id {
		t.Errorf("Expected ID %s, got %s", id, work.GetID())
	}

	if work.GetType() != workType {
		t.Errorf("Expected type %s, got %s", workType, work.GetType())
	}

	if work.GetStatus() != WorkStatusPending {
		t.Errorf("Expected status %s, got %s", WorkStatusPending, work.GetStatus())
	}

	if work.GetPriority() != WorkPriorityNormal {
		t.Errorf("Expected priority %d, got %d", WorkPriorityNormal, work.GetPriority())
	}

	if work.GetMetadata().Name != name {
		t.Errorf("Expected name %s, got %s", name, work.GetMetadata().Name)
	}
}

func TestWorkSetStatus(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")
	originalTime := work.Metadata.UpdatedAt

	time.Sleep(1 * time.Millisecond) // Ensure time difference

	newWork := work.SetStatus(WorkStatusExecuting)

	if newWork.GetStatus() != WorkStatusExecuting {
		t.Errorf("Expected status %s, got %s", WorkStatusExecuting, newWork.GetStatus())
	}

	if newWork.Metadata.UpdatedAt.Equal(originalTime) {
		t.Error("UpdatedAt should be updated when status changes")
	}

	// Original work should remain unchanged (immutability)
	if work.GetStatus() != WorkStatusPending {
		t.Error("Original work should remain unchanged")
	}
}

func TestWorkSetInput(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")
	testInput := map[string]interface{}{"key": "value"}

	newWork := work.SetInput(testInput)

	if newWork.GetInput() == nil {
		t.Error("Input should be set")
	}

	// Original work should remain unchanged (immutability)
	if work.GetInput() != nil {
		t.Error("Original work should remain unchanged")
	}
}

func TestWorkSetOutput(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")
	testOutput := map[string]interface{}{"result": "success"}

	newWork := work.SetOutput(testOutput)

	if newWork.GetOutput() == nil {
		t.Error("Output should be set")
	}

	// Original work should remain unchanged (immutability)
	if work.GetOutput() != nil {
		t.Error("Original work should remain unchanged")
	}
}

func TestWorkSetError(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")
	errorMsg := "test error"

	newWork := work.SetError(errorMsg)

	if newWork.GetError() != errorMsg {
		t.Errorf("Expected error %s, got %s", errorMsg, newWork.GetError())
	}

	// Original work should remain unchanged (immutability)
	if work.GetError() != "" {
		t.Error("Original work should remain unchanged")
	}
}

func TestWorkSetCompensationWorkID(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")
	compensationID := WorkID("compensation-work")

	newWork := work.SetCompensationWorkID(compensationID)

	if newWork.GetCompensationWorkID() == nil || *newWork.GetCompensationWorkID() != compensationID {
		t.Errorf("Expected compensation work ID %s", compensationID)
	}

	// Original work should remain unchanged (immutability)
	if work.GetCompensationWorkID() != nil {
		t.Error("Original work should remain unchanged")
	}
}

func TestWorkMarkStarted(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")

	startedWork := work.MarkStarted()

	if startedWork.GetStatus() != WorkStatusExecuting {
		t.Errorf("Expected status %s, got %s", WorkStatusExecuting, startedWork.GetStatus())
	}

	if startedWork.GetMetadata().StartedAt == nil {
		t.Error("StartedAt should be set")
	}
}

func TestWorkMarkCompleted(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")
	output := "test output"

	completedWork := work.MarkCompleted(output)

	if completedWork.GetStatus() != WorkStatusCompleted {
		t.Errorf("Expected status %s, got %s", WorkStatusCompleted, completedWork.GetStatus())
	}

	if completedWork.GetOutput() != output {
		t.Errorf("Expected output %s, got %v", output, completedWork.GetOutput())
	}

	if completedWork.GetMetadata().CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestWorkMarkFailed(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")
	errorMsg := "test error"

	failedWork := work.MarkFailed(errorMsg)

	if failedWork.GetStatus() != WorkStatusFailed {
		t.Errorf("Expected status %s, got %s", WorkStatusFailed, failedWork.GetStatus())
	}

	if failedWork.GetError() != errorMsg {
		t.Errorf("Expected error %s, got %s", errorMsg, failedWork.GetError())
	}

	if failedWork.GetMetadata().CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestWorkIsExecutable(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")

	// Test pending status
	if !work.IsExecutable() {
		t.Error("Pending work should be executable")
	}

	// Test scheduled status
	scheduledWork := work.SetStatus(WorkStatusScheduled)
	if !scheduledWork.IsExecutable() {
		t.Error("Scheduled work should be executable")
	}

	// Test retrying status
	retryingWork := work.SetStatus(WorkStatusRetrying)
	if !retryingWork.IsExecutable() {
		t.Error("Retrying work should be executable")
	}

	// Test executing status
	executingWork := work.SetStatus(WorkStatusExecuting)
	if executingWork.IsExecutable() {
		t.Error("Executing work should not be executable")
	}

	// Test completed status
	completedWork := work.SetStatus(WorkStatusCompleted)
	if completedWork.IsExecutable() {
		t.Error("Completed work should not be executable")
	}
}

func TestWorkIsCompleted(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")

	if work.IsCompleted() {
		t.Error("New work should not be completed")
	}

	completedWork := work.SetStatus(WorkStatusCompleted)
	if !completedWork.IsCompleted() {
		t.Error("Work with completed status should be completed")
	}
}

func TestWorkIsFailed(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")

	if work.IsFailed() {
		t.Error("New work should not be failed")
	}

	failedWork := work.SetStatus(WorkStatusFailed)
	if !failedWork.IsFailed() {
		t.Error("Work with failed status should be failed")
	}
}

func TestWorkRequiresCompensation(t *testing.T) {
	work := NewWork("test", WorkTypeTask, "Test")
	compensationID := WorkID("compensation-work")

	// Work without compensation ID should not require compensation
	completedWork := work.SetStatus(WorkStatusCompleted)
	if completedWork.RequiresCompensation() {
		t.Error("Work without compensation ID should not require compensation")
	}

	// Work with compensation ID but not completed should not require compensation
	workWithCompensation := work.SetCompensationWorkID(compensationID)
	if workWithCompensation.RequiresCompensation() {
		t.Error("Non-completed work should not require compensation")
	}

	// Completed work with compensation ID should require compensation
	completedWorkWithCompensation := workWithCompensation.SetStatus(WorkStatusCompleted)
	if !completedWorkWithCompensation.RequiresCompensation() {
		t.Error("Completed work with compensation ID should require compensation")
	}
}

func TestWorkClone(t *testing.T) {
	original := NewWork("test", WorkTypeTask, "Test")
	original.Metadata.Tags = []string{"tag1", "tag2"}
	original.Metadata.Properties["key"] = "value"
	original.Configuration.Parameters["param"] = "value"
	original.Configuration.Environment["env"] = "value"
	compensationID := WorkID("compensation")
	original = original.SetCompensationWorkID(compensationID)

	cloned := original.Clone()

	// Verify clone has same values
	if cloned.GetID() != original.GetID() {
		t.Error("Cloned work should have same ID")
	}

	if *cloned.GetCompensationWorkID() != *original.GetCompensationWorkID() {
		t.Error("Cloned work should have same compensation work ID")
	}

	// Verify independence (modify clone)
	cloned.Metadata.Tags[0] = "modified"
	cloned.Metadata.Properties["key"] = "modified"
	cloned.Configuration.Parameters["param"] = "modified"
	cloned.Configuration.Environment["env"] = "modified"

	if original.Metadata.Tags[0] == "modified" {
		t.Error("Original work tags should not be affected by clone modification")
	}

	if original.Metadata.Properties["key"] == "modified" {
		t.Error("Original work properties should not be affected by clone modification")
	}

	if original.Configuration.Parameters["param"] == "modified" {
		t.Error("Original work parameters should not be affected by clone modification")
	}

	if original.Configuration.Environment["env"] == "modified" {
		t.Error("Original work environment should not be affected by clone modification")
	}
}

func TestWorkValidate(t *testing.T) {
	// Valid work
	validWork := NewWork("test", WorkTypeTask, "Test")
	if err := validWork.Validate(); err != nil {
		t.Errorf("Valid work should not return error: %v", err)
	}

	// Invalid works
	invalidWorks := []Work{
		{ID: "", Type: WorkTypeTask, Status: WorkStatusPending, Metadata: WorkMetadata{Name: "Test"}, Configuration: WorkConfiguration{TimeoutSeconds: 300, RetryCount: 3, RetryDelaySeconds: 5}},
		{ID: "test", Type: "", Status: WorkStatusPending, Metadata: WorkMetadata{Name: "Test"}, Configuration: WorkConfiguration{TimeoutSeconds: 300, RetryCount: 3, RetryDelaySeconds: 5}},
		{ID: "test", Type: WorkTypeTask, Status: "", Metadata: WorkMetadata{Name: "Test"}, Configuration: WorkConfiguration{TimeoutSeconds: 300, RetryCount: 3, RetryDelaySeconds: 5}},
		{ID: "test", Type: WorkTypeTask, Status: WorkStatusPending, Metadata: WorkMetadata{Name: ""}, Configuration: WorkConfiguration{TimeoutSeconds: 300, RetryCount: 3, RetryDelaySeconds: 5}},
		{ID: "test", Type: WorkTypeTask, Status: WorkStatusPending, Metadata: WorkMetadata{Name: "Test"}, Configuration: WorkConfiguration{TimeoutSeconds: 0, RetryCount: 3, RetryDelaySeconds: 5}},
		{ID: "test", Type: WorkTypeTask, Status: WorkStatusPending, Metadata: WorkMetadata{Name: "Test"}, Configuration: WorkConfiguration{TimeoutSeconds: 300, RetryCount: -1, RetryDelaySeconds: 5}},
		{ID: "test", Type: WorkTypeTask, Status: WorkStatusPending, Metadata: WorkMetadata{Name: "Test"}, Configuration: WorkConfiguration{TimeoutSeconds: 300, RetryCount: 3, RetryDelaySeconds: -1}},
	}

	for i, work := range invalidWorks {
		if err := work.Validate(); err == nil {
			t.Errorf("Invalid work %d should return error", i)
		}
	}
}
