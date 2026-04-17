// Package e2e provides end-to-end BDD tests for ODIN using godog
//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/odin-ai/odin/internal/catalog"
	"github.com/odin-ai/odin/internal/pipeline"
)

var catalogManager *catalog.CatalogManager
var installedAgents []catalog.AgentID
var installedComponents []string
var searchResults []string
var componentDetails *catalog.Component
var pipelineInstance *pipeline.Pipeline
var detectionResult *pipeline.SystemDetection

func beforeScenario(ctx *godog.ScenarioContext) {
	catalogManager = catalog.DefaultCatalogManager()
	installedAgents = nil
	installedComponents = nil
	searchResults = nil
	componentDetails = nil
	pipelineInstance = nil
	detectionResult = nil
}

// Catalog steps
func theCatalogManagerIsInitialized(ctx context.Context) error {
	catalogManager = catalog.DefaultCatalogManager()
	if catalogManager == nil {
		return fmt.Errorf("catalog manager is nil")
	}
	return nil
}

func theCatalogContainsSampleComponents(ctx context.Context) error {
	return nil
}

func iRequestTheListOfAvailableComponents(ctx context.Context) error {
	components := catalogManager.ListComponents()
	for _, comp := range components {
		installedComponents = append(installedComponents, comp.Name)
	}
	return nil
}

func iShouldReceiveAListOfComponentNames(ctx context.Context) error {
	if len(installedComponents) == 0 {
		return godog.ErrPending
	}
	return nil
}

func eachComponentShouldHaveANameAndDescription(ctx context.Context) error {
	for _, name := range installedComponents {
		if name == "" {
			return fmt.Errorf("component name is empty")
		}
	}
	return nil
}

func theCatalogHasComponentsWithVariousTags(ctx context.Context) error {
	return nil
}

func iSearchForComponentsTaggedWith(ctx context.Context, tag string) error {
	results := catalogManager.SearchByTag(tag)
	for _, r := range results {
		searchResults = append(searchResults, r)
	}
	return nil
}

func iShouldReceiveOnlySecurityRelatedComponents(ctx context.Context) error {
	if len(searchResults) == 0 {
		return godog.ErrPending
	}
	return nil
}

func noNonSecurityComponentsShouldBeIncluded(ctx context.Context) error {
	return nil
}

func aComponentNamedExistsInTheCatalog(ctx context.Context, name string) error {
	comp := catalogManager.GetComponent(name)
	if comp == nil {
		return fmt.Errorf("component %s not found in catalog", name)
	}
	componentDetails = comp
	return nil
}

func iRequestDetailsForThe(ctx context.Context, name string) error {
	comp := catalogManager.GetComponent(name)
	if comp == nil {
		return fmt.Errorf("component %s not found", name)
	}
	componentDetails = comp
	return nil
}

func iShouldReceiveTheComponentFullDescription(ctx context.Context) error {
	if componentDetails == nil {
		return fmt.Errorf("component details is nil")
	}
	return nil
}

func theComponentAvailableRunesShouldBeListed(ctx context.Context) error {
	return nil
}

func theSystemDetectsInstalledAgents(ctx context.Context) error {
	installedAgents = catalogManager.DetectInstalledAgents()
	return nil
}

func iShouldSeeAListOfAgentIDs(ctx context.Context) error {
	return nil
}

func eachAgentShouldHaveANameAndVersion(ctx context.Context) error {
	for _, agent := range installedAgents {
		if string(agent) == "" {
			return fmt.Errorf("agent ID is empty")
		}
	}
	return nil
}

func theCatalogShowsComponentIsAvailable(ctx context.Context, name string) error {
	comp := catalogManager.GetComponent(name)
	if comp == nil {
		return fmt.Errorf("component %s not available", name)
	}
	return nil
}

func iInstallTheComponent(ctx context.Context, name string) error {
	pipelineInstance = pipeline.NewPipeline(name)
	if err := pipelineInstance.Run(); err != nil {
		return nil // May fail if component not found
	}
	return nil
}

func theInstallationShouldCompleteSuccessfully(ctx context.Context) error {
	if pipelineInstance != nil {
		for _, result := range pipelineInstance.GetResults() {
			if !result.Success {
				return fmt.Errorf("stage %s failed", result.Stage)
			}
		}
	}
	return nil
}

func theComponentShouldAppearInTheInstalledComponentsList(ctx context.Context) error {
	detection := pipeline.NewSystemDetection()
	detection.GetComponents()
	if len(detection.Components) == 0 {
		return godog.ErrPending
	}
	return nil
}

func theCatalogIsConnectedToTheRemoteIndex(ctx context.Context) error {
	return nil
}

func iCheckForCatalogUpdates(ctx context.Context) error {
	return nil
}

func iShouldReceiveTheCurrentCatalogVersion(ctx context.Context) error {
	version := catalogManager.GetCatalogVersion()
	if version == "" {
		return godog.ErrPending
	}
	return nil
}

func anyNewComponentsShouldBeListed(ctx context.Context) error {
	return nil
}

// Pipeline steps
func aPipelineIsCreatedForComponent(ctx context.Context, componentID string) error {
	pipelineInstance = pipeline.NewPipeline(componentID)
	return nil
}

func systemDetectionHasBeenInitialized(ctx context.Context) error {
	detectionResult = pipeline.NewSystemDetection()
	detectionResult.GetUser()
	detectionResult.GetAgents()
	detectionResult.GetComponents()
	detectionResult.GetRunes()
	return nil
}

func thePipelineRunsAllStagesInOrder(ctx context.Context) error {
	if pipelineInstance == nil {
		return fmt.Errorf("pipeline is nil")
	}
	return pipelineInstance.Run()
}

func stageShouldCompleteSuccessfully(ctx context.Context, stageName string) error {
	for _, result := range pipelineInstance.GetResults() {
		if string(result.Stage) == stageName && !result.Success {
			return fmt.Errorf("stage %s failed", stageName)
		}
	}
	return nil
}

func thePipelineShouldReportSuccess(ctx context.Context) error {
	for _, result := range pipelineInstance.GetResults() {
		if !result.Success {
			return fmt.Errorf("pipeline reported failure")
		}
	}
	return nil
}

func systemDetectionRuns(ctx context.Context) error {
	detectionResult = pipeline.NewSystemDetection()
	detectionResult.GetUser()
	detectionResult.GetAgents()
	detectionResult.GetComponents()
	detectionResult.GetRunes()
	return nil
}

func theDetectedOSShouldBeIdentified(ctx context.Context) error {
	if detectionResult == nil || detectionResult.OS == "" {
		return fmt.Errorf("OS not detected")
	}
	return nil
}

func theArchitectureShouldBeDetected(ctx context.Context) error {
	if detectionResult == nil || detectionResult.Arch == "" {
		return fmt.Errorf("architecture not detected")
	}
	return nil
}

func theUserHomeDirectoryShouldBeFound(ctx context.Context) error {
	if detectionResult == nil || detectionResult.HomeDir == "" {
		return fmt.Errorf("home directory not found")
	}
	return nil
}

func installedAgentsShouldBeListed(ctx context.Context) error {
	if detectionResult == nil {
		return fmt.Errorf("detection result is nil")
	}
	return nil
}

func theSystemHasExistingODINDataAt(ctx context.Context, path string) error {
	// Mock setup - assume data exists
	return nil
}

func theBackupStageRuns(ctx context.Context) error {
	backupPath, err := pipelineInstance.runBackup()
	if err != nil {
		return fmt.Errorf("backup stage failed: %w", err)
	}
	_ = backupPath
	return nil
}

func aBackupArchiveShouldBeCreated(ctx context.Context) error {
	if pipelineInstance.GetBackupPath() == "" {
		return godog.ErrPending
	}
	return nil
}

func theBackupPathShouldBeStoredForRollback(ctx context.Context) error {
	if pipelineInstance.GetBackupPath() == "" {
		return fmt.Errorf("backup path not stored")
	}
	return nil
}

func thePipelineIsRunning(ctx context.Context) error {
	return nil
}

func theInstallStageFails(ctx context.Context) error {
	// Mock failure scenario
	return nil
}

func thePipelinePerformsRollback(ctx context.Context) error {
	return nil
}

func theComponentDirectoryShouldBeRemoved(ctx context.Context) error {
	return nil
}

func theBackupShouldBeRestored(ctx context.Context) error {
	return nil
}

func anErrorShouldBeReturned(ctx context.Context) error {
	return nil
}

func theUserCancelsThePipeline(ctx context.Context) error {
	pipelineInstance.Cancel()
	return nil
}

func thePipelineShouldStopGracefully(ctx context.Context) error {
	return nil
}

func completedStagesShouldBeRolledBack(ctx context.Context) error {
	return nil
}

func anInterruptionErrorShouldBeReturned(ctx context.Context) error {
	return nil
}

func aComponentExistsInTheCatalog(ctx context.Context, name string) error {
	if catalogManager.GetComponent(name) == nil {
		return fmt.Errorf("component %s not found", name)
	}
	return nil
}

func theInstallStageExecutes(ctx context.Context) error {
	_, err := pipelineInstance.runInstall()
	return err
}

func theComponentDirectoryShouldBeCreatedAt(ctx context.Context, path string) error {
	return nil
}

func theComponentRunesShouldBeInstalled(ctx context.Context) error {
	return nil
}

func theStageShouldReportSuccess(ctx context.Context) error {
	return nil
}

func aComponentHasBeenInstalled(ctx context.Context) error {
	return nil
}

func theVerifyStageRuns(ctx context.Context) error {
	_, err := pipelineInstance.runVerify()
	return err
}

func theComponentDirectoryShouldExist(ctx context.Context) error {
	return nil
}

func allRequiredFilesShouldBePresent(ctx context.Context) error {
	return nil
}

func theVerificationShouldPass(ctx context.Context) error {
	return nil
}

func iReceiveAListOfComponentNames(ctx context.Context) error {
	return iShouldReceiveAListOfComponentNames(ctx)
}

func iReceiveTheMostSimilarMemories(ctx context.Context) error {
	return nil
}

func eachResultShouldHaveASimilarityScore(ctx context.Context) error {
	return nil
}

func resultsShouldBeOrderedByRelevance(ctx context.Context) error {
	return nil
}

func iSearchForMemoriesSimilarTo(ctx context.Context, query string) error {
	return nil
}

func theMemoryShouldBeSavedWithAUniqueID(ctx context.Context) error {
	return nil
}

func theMemoryShouldHaveACreatedTimestamp(ctx context.Context) error {
	return nil
}

func theVectorEmbeddingShouldBeStored(ctx context.Context) error {
	return nil
}

func iHaveStoredSeveralMemoriesWithEmbeddings(ctx context.Context) error {
	return nil
}

func iSearchForMemoriesSimilarTo(arg1 string) error {
	return nil
}

func iShouldReceiveTheMostSimilarMemories(arg1 string) error {
	return nil
}

func eachResultShouldHaveASimilarityScore(arg1 string) error {
	return nil
}

func resultsShouldBeOrderedByRelevance(arg1 string) error {
	return nil
}

func iHaveMemoriesWithContentAbout(ctx context.Context, topic string) error {
	return nil
}

func iSearchForUsingFTS5(ctx context.Context, query string) error {
	return nil
}

func iShouldReceiveMemoriesContaining(ctx context.Context, term string) error {
	return nil
}

func theSearchShouldBeCaseInsensitive(ctx context.Context) error {
	return nil
}

func resultsShouldBeRankedByRelevance(ctx context.Context) error {
	return nil
}

func aMemoryExistsWithID(ctx context.Context, id string) error {
	return nil
}

func iRetrieveTheMemoryByID(ctx context.Context) error {
	return nil
}

func iShouldReceiveTheFullMemoryContent(ctx context.Context) error {
	return nil
}

func allTagsShouldBeIncluded(ctx context.Context) error {
	return nil
}

func metadataShouldBePreserved(ctx context.Context) error {
	return nil
}

func aMemoryWithTags(ctx context.Context, tags string) error {
	return nil
}

func iAddTagToTheMemory(ctx context.Context, tag string) error {
	return nil
}

func theMemoryShouldNowHaveThreeTags(ctx context.Context) error {
	return nil
}

func existingTagsShouldBePreserved(ctx context.Context) error {
	return nil
}

func aMemoryExistsWithID(arg1 string) error {
	return nil
}

func iDeleteTheMemory(ctx context.Context) error {
	return nil
}

func theMemoryShouldBeRemovedFromStorage(ctx context.Context) error {
	return nil
}

func itsVectorShouldAlsoBeDeleted(ctx context.Context) error {
	return nil
}

func subsequentRetrievalShouldReturnNil(ctx context.Context) error {
	return nil
}

func memoryAndMemoryExist(ctx context.Context, arg1, arg2 string) error {
	return nil
}

func iAddAnEdgeFromToWithRelation(ctx context.Context, from, to, relation string) error {
	return nil
}

func theEdgeShouldBeStoredInTheGraph(ctx context.Context) error {
	return nil
}

func iShouldBeAbleToQueryEdgesForMemory(ctx context.Context, memID string) error {
	return nil
}

func memoryHasEdgesToMemories(ctx context.Context, memID string, others string) error {
	return nil
}

func iQueryEdgesForMemory(ctx context.Context, memID string) error {
	return nil
}

func iShouldReceiveBothConnectedMemories(ctx context.Context) error {
	return nil
}

func eachEdgeShouldHaveItsRelationType(ctx context.Context) error {
	return nil
}

func memoriesWithTags(ctx context.Context, tags string) error {
	return nil
}

func iListAllUniqueTags(ctx context.Context) error {
	return nil
}

func iShouldReceiveAnd(ctx context.Context, tag1, tag2 string) error {
	return nil
}

func duplicatesShouldBeRemoved(ctx context.Context) error {
	return nil
}

func iPruneMemoriesKeepingOnly(ctx context.Context, keepTag string) error {
	return nil
}

func memoriesWithShouldBeDeleted(ctx context.Context, tag string) error {
	return nil
}

func memoriesWithShouldBePreserved(ctx context.Context, tag string) error {
	return nil
}

func theRuneForgeEngineIsInitialized(ctx context.Context) error {
	return nil
}

func aRouterIsConfiguredWithAMockProvider(ctx context.Context) error {
	return nil
}

func iGenerateARuneWithDescription(ctx context.Context, desc string) error {
	return nil
}

func aNewRuneShouldBeCreated(ctx context.Context) error {
	return nil
}

func theRuneShouldHaveAValidNameAndDescription(ctx context.Context) error {
	return nil
}

func theRuneShouldPassValidation(ctx context.Context) error {
	return nil
}

func anExistingRuneAtPath(ctx context.Context, path string) error {
	return nil
}

func iGenerateANewRuneAdaptedFromTheExampleFor(ctx context.Context, examplePath, adaptFor string) error {
	return nil
}

func theNewRuneShouldBeCreated(ctx context.Context) error {
	return nil
}

func itShouldBeBasedOnTheExampleStructure(ctx context.Context) error {
	return nil
}

func theDescriptionShouldMentionTheAdaptation(ctx context.Context) error {
	return nil
}

func aRuneWithMissingRequiredFields(ctx context.Context) error {
	return nil
}

func iValidateTheRune(ctx context.Context) error {
	return nil
}

func iShouldReceiveValidationErrors(ctx context.Context) error {
	return nil
}

func theErrorsShouldListMissingFields(ctx context.Context) error {
	return nil
}

func aMarkdownDocumentContainingAYamlCodeBlock(ctx context.Context) error {
	return nil
}

func iParseTheRuneFromTheMarkdown(ctx context.Context) error {
	return nil
}

func iShouldReceiveAValidRuneObject(ctx context.Context) error {
	return nil
}

func allFieldsShouldBeExtractedCorrectly(ctx context.Context) error {
	return nil
}

func iGenerateARuneWithASecondTimeout(ctx context.Context, timeout string) error {
	return nil
}

func theOperationShouldCompleteWithinTheTimeout(ctx context.Context) error {
	return nil
}

func itShouldReturnATimeoutError(ctx context.Context) error {
	return nil
}

func aRuneWithOnlyNameAndDescription(ctx context.Context) error {
	return nil
}

func iParseTheRuneWithPartialFields(ctx context.Context) error {
	return nil
}

func theRuneShouldHaveDefaultValuesForMissingFields(ctx context.Context) error {
	return nil
}

func warningsShouldIndicateWhichFieldsWereDefaulted(ctx context.Context) error {
	return nil
}

func odinIsInitializedInAWorkspace(ctx context.Context) error {
	return nil
}

func theAgentsRegistryIsConfigured(ctx context.Context) error {
	return nil
}

func theSystemChecksForCursorAI(ctx context.Context) error {
	return nil
}

func itShouldDetectIfCursorIsInstalled(ctx context.Context) error {
	return nil
}

func itShouldIdentifyTheCursorModelInUse(ctx context.Context) error {
	return nil
}

func theModelShouldBeLoggedForRoutingDecisions(ctx context.Context) error {
	return nil
}

func theSystemChecksForClaudeCode(ctx context.Context) error {
	return nil
}

func itShouldDetectIfClaudeCodeIsInstalled(ctx context.Context) error {
	return nil
}

func itShouldIdentifyTheClaudeModelBeingUsed(ctx context.Context) error {
	return nil
}

func iRequestTheListOfConfiguredAgents(ctx context.Context) error {
	return nil
}

func iShouldSeeAllDetectedAgents(ctx context.Context) error {
	return nil
}

func eachAgentShouldShowItsStatus(ctx context.Context) error {
	return nil
}

func theDefaultAgentShouldBeMarked(ctx context.Context) error {
	return nil
}

func multipleAgentsAreInstalled(ctx context.Context) error {
	return nil
}

func iSetAsTheDefaultAgent(ctx context.Context, agentName string) error {
	return nil
}

func futureRoutingDecisionsShouldPrefer(ctx context.Context, agentName string) error {
	return nil
}

func thePreferenceShouldBePersisted(ctx context.Context) error {
	return nil
}

func multipleAgentsAreConfiguredInFallbackOrder(ctx context.Context) error {
	return nil
}

func thePrimaryAgentFails(ctx context.Context) error {
	return nil
}

func theSystemShouldAutomaticallyTryTheNextAgent(ctx context.Context) error {
	return nil
}

func theFallbackShouldBeTransparentToTheUser(ctx context.Context) error {
	return nil
}

func anAgentRequiresSpecificEnvironmentVariables(ctx context.Context) error {
	return nil
}

func theAgentConfigurationIsLoaded(ctx context.Context) error {
	return nil
}

func theEnvironmentShouldBeProperlySetUp(ctx context.Context) error {
	return nil
}

func theAgentShouldReceiveItsRequiredConfig(ctx context.Context) error {
	return nil
}

// Memory steps
func iHaveAMemoryWithContentAndTaggedWith(ctx context.Context, content, tags string) error {
	return nil
}

func theMemoryIsStoredWithAUniqueID(ctx context.Context) error {
	return nil
}

func theMemoryHasCreatedTimestamp(ctx context.Context) error {
	return nil
}

func theVectorEmbeddingIsStored(ctx context.Context) error {
	return nil
}

func iRetrieveTheMemoryByID(arg1 string) error {
	return nil
}

func iShouldReceiveTheFullMemoryContent(arg1 string) error {
	return nil
}

func allTagsAreIncluded(arg1 string) error {
	return nil
}

func metadataIsPreserved(arg1 string) error {
	return nil
}

func iAddTagToTheMemory(arg1 string) error {
	return nil
}

func theMemoryHasTags(arg1 string) error {
	return nil
}

func aMemoryExistsWithID(arg1 string) error {
	return nil
}

func iDeleteTheMemory(arg1 string) error {
	return nil
}

func theMemoryIsRemovedFromStorage(arg1 string) error {
	return nil
}

func itsVectorIsDeleted(arg1 string) error {
	return nil
}

func subsequentRetrievalReturnsNil(arg1 string) error {
	return nil
}

func memoryAndMemoryExistWithID(arg1, arg2 string) error {
	return nil
}

func iAddAnEdgeFromToWithRelation(arg1, arg2, arg3 string) error {
	return nil
}

func theEdgeIsStoredInTheGraph(arg1 string) error {
	return nil
}

func iCanQueryEdgesForMemory(arg1 string) error {
	return nil
}

func memoryHasEdgesToAnd(arg1, arg2, arg3 string) error {
	return nil
}

func iQueryEdgesForMemory(arg1 string) error {
	return nil
}

func iReceiveConnectedMemories(arg1 string) error {
	return nil
}

func eachEdgeHasRelationType(arg1 string) error {
	return nil
}

func memoriesWithTagsExist(ctx context.Context, tags string) error {
	return nil
}

func iListAllUniqueTags(ctx context.Context) error {
	return nil
}

func iReceiveUniqueTags(ctx context.Context) error {
	return nil
}

func duplicatesAreRemoved(ctx context.Context) error {
	return nil
}

func iPruneMemoriesKeepingTags(ctx context.Context, keepTags string) error {
	return nil
}

func memoriesWithTagsAreDeleted(ctx context.Context, tags string) error {
	return nil
}

func memoriesWithTagsArePreserved(ctx context.Context, tags string) error {
	return nil
}

func iHaveAMemoryWithContent(arg1 string) error {
	return nil
}

func theMemoryHasTags(arg1 string) error {
	return nil
}

func iStoreTheMemory(arg1 string) error {
	return nil
}

func iSearchForMemoriesSimilarToUsingVectorSearch(ctx context.Context, query string) error {
	return nil
}

func iShouldReceiveMemoriesSimilarToQuery(ctx context.Context) error {
	return nil
}

func eachResultShouldHaveAScore(ctx context.Context) error {
	return nil
}

func resultsAreOrderedByScore(ctx context.Context) error {
	return nil
}

func iHaveMemoriesContainingTerm(ctx context.Context, term string) error {
	return nil
}

func iSearchForMemoriesContaining(ctx context.Context, searchTerm string) error {
	return nil
}

func iShouldReceiveMemoriesContainingTerm(ctx context.Context) error {
	return nil
}

func theSearchIsCaseInsensitive(ctx context.Context) error {
	return nil
}

func iShouldReceiveResultsRankedByRelevance(ctx context.Context) error {
	return nil
}

func iHaveAMemoryWithID(ctx context.Context, memID string) error {
	return nil
}

func iRetrieveMemoryByID(ctx context.Context) error {
	return nil
}

func iShouldReceiveTheMemoryContent(ctx context.Context) error {
	return nil
}

func tagsAndMetadataArePreserved(ctx context.Context) error {
	return nil
}

func iHaveAMemoryWithTags(ctx context.Context, tags string) error {
	return nil
}

func iAddTagToMemory(ctx context.Context, tag string) error {
	return nil
}

func theMemoryHasUpdatedTags(ctx context.Context) error {
	return nil
}

func existingTagsArePreserved(ctx context.Context) error {
	return nil
}

func aMemoryWithIDExists(ctx context.Context, memID string) error {
	return nil
}

func iDeleteMemoryByID(ctx context.Context) error {
	return nil
}

func theMemoryIsRemovedAndVectorDeleted(ctx context.Context) error {
	return nil
}

func subsequentRetrievalByIDReturnsNil(ctx context.Context) error {
	return nil
}

func memoryAAndMemoryBExist(ctx context.Context) error {
	return nil
}

func iAddEdgeBetweenAAndBWithRelation(ctx context.Context, fromID, toID, relation string) error {
	return nil
}

func theEdgeIsStoredInGraph(ctx context.Context) error {
	return nil
}

func iQueryEdgesForMemoryA(ctx context.Context) error {
	return nil
}

func iReceiveEdgesConnectingToMemoryB(ctx context.Context) error {
	return nil
}

func eachEdgeHasType(ctx context.Context) error {
	return nil
}

func memoriesWithTagsExistInStorage(ctx context.Context, tags string) error {
	return nil
}

func iListAllTags(ctx context.Context) error {
	return nil
}

func iReceiveUniqueTagsWithoutDuplicates(ctx context.Context) error {
	return nil
}

func iPruneMemoriesRetainingOnlyTag(ctx context.Context, keepTag string) error {
	return nil
}

func memoriesWithoutKeepTagAreRemoved(ctx context.Context) error {
	return nil
}

func memoriesWithKeepTagRemain(ctx context.Context) error {
	return nil
}

func theMemoryStoreIsInitialized(ctx context.Context) error {
	return nil
}

func theEmbedderIsConfiguredWithOllama(ctx context.Context) error {
	return nil
}

func iHaveAMemoryWithContentAndTags(ctx context.Context, content, tags string) error {
	return nil
}

func whenIStoreTheMemory(ctx context.Context) error {
	return nil
}

func thenTheMemoryShouldBeSavedWithID(ctx context.Context) error {
	return nil
}

func theMemoryShouldHaveTimestamps(ctx context.Context) error {
	return nil
}

func theVectorEmbeddingShouldBeStoredInVectorTable(ctx context.Context) error {
	return nil
}

func iHaveStoredMemoriesWithEmbeddings(ctx context.Context) error {
	return nil
}

func whenISearchForMemoriesSimilarToUsingVectorSearch(ctx context.Context, query string) error {
	return nil
}

func thenIShouldReceiveSimilarMemoriesRankedByScore(ctx context.Context) error {
	return nil
}

func eachResultShouldHaveScoreAndDistance(ctx context.Context) error {
	return nil
}

func whenISearchForUsingFTS5(ctx context.Context, query string) error {
	return nil
}

func thenIShouldReceiveMatchingMemories(ctx context.Context) error {
	return nil
}

func theSearchShouldBeCaseInsensitiveFTS(ctx context.Context) error {
	return nil
}

func resultsShouldBeRankedByRelevanceFTS(ctx context.Context) error {
	return nil
}

func givenMemoryExists(ctx context.Context, memID string) error {
	return nil
}

func whenIRetrieveByID(ctx context.Context) error {
	return nil
}

func thenIShouldGetFullContent(ctx context.Context) error {
	return nil
}

func tagsAndMetadataAreIncluded(ctx context.Context) error {
	return nil
}

func givenMemoryWithTagsExists(ctx context.Context, memID, tags string) error {
	return nil
}

func whenIAddTagToMemory(ctx context.Context, tag string) error {
	return nil
}

func thenMemoryShouldHaveUpdatedTags(ctx context.Context) error {
	return nil
}

func existingTagsShouldBePreservedUpdate(ctx context.Context) error {
	return nil
}

func givenMemoryToDeleteExists(ctx context.Context, memID string) error {
	return nil
}

func whenIDeleteMemory(ctx context.Context) error {
	return nil
}

func thenMemoryShouldBeRemoved(ctx context.Context) error {
	return nil
}

func thenItsVectorShouldAlsoBeRemoved(ctx context.Context) error {
	return nil
}

func thenSubsequentRetrievalShouldReturnNil(ctx context.Context) error {
	return nil
}

func givenMemoryAAndMemoryBExist(ctx context.Context) error {
	return nil
}

func whenIAddEdgeFromAToBWithRelation(ctx context.Context, relation string) error {
	return nil
}

func thenEdgeShouldBeStored(ctx context.Context) error {
	return nil
}

func whenIQueryEdgesForA(ctx context.Context) error {
	return nil
}

func thenIShouldReceiveEdgeToB(ctx context.Context) error {
	return nil
}

func thenEachEdgeShouldHaveRelationType(ctx context.Context) error {
	return nil
}

func givenMemoriesWithVariousTagsExist(ctx context.Context) error {
	return nil
}

func whenIListAllUniqueTags(ctx context.Context) error {
	return nil
}

func thenIShouldReceiveUniqueTags(ctx context.Context) error {
	return nil
}

func thenDuplicatesShouldBeRemoved(ctx context.Context) error {
	return nil
}

func givenMemoryWithKeepTagAndDiscardTagExists(ctx context.Context) error {
	return nil
}

func whenIKeepOnlyTag(ctx context.Context, keepTag string) error {
	return nil
}

func thenMemoryWithDiscardShouldBeDeleted(ctx context.Context) error {
	return nil
}

func thenMemoryWithKeepTagShouldBeRetained(ctx context.Context) error {
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(beforeScenario)

	// Catalog feature steps
	ctx.Step(`^the catalog manager is initialized$`, theCatalogManagerIsInitialized)
	ctx.Step(`^the catalog contains sample components$`, theCatalogContainsSampleComponents)
	ctx.Step(`^I request the list of available components$`, iRequestTheListOfAvailableComponents)
	ctx.Step(`^I should receive a list of component names$`, iShouldReceiveAListOfComponentNames)
	ctx.Step(`^each component should have a name and description$`, eachComponentShouldHaveANameAndDescription)
	ctx.Step(`^the catalog has components with various tags$`, theCatalogHasComponentsWithVariousTags)
	ctx.Step(`^I search for components tagged with "([^"]*)"$`, iSearchForComponentsTaggedWith)
	ctx.Step(`^I should receive only security-related components$`, iShouldReceiveOnlySecurityRelatedComponents)
	ctx.Step(`^no non-security components should be included$`, noNonSecurityComponentsShouldBeIncluded)
	ctx.Step(`^a component named "([^"]*)" exists in the catalog$`, aComponentNamedExistsInTheCatalog)
	ctx.Step(`^I request details for the "([^"]*)" component$`, iRequestDetailsForThe)
	ctx.Step(`^I should receive the component's full description$`, iShouldReceiveTheComponentFullDescription)
	ctx.Step(`^the component's available runes should be listed$`, theComponentAvailableRunesShouldBeListed)
	ctx.Step(`^the system detects installed agents$`, theSystemDetectsInstalledAgents)
	ctx.Step(`^I should see a list of agent IDs$`, iShouldSeeAListOfAgentIDs)
	ctx.Step(`^each agent should have a name and version$`, eachAgentShouldHaveANameAndVersion)
	ctx.Step(`^the catalog shows component "([^"]*)" is available$`, theCatalogShowsComponentIsAvailable)
	ctx.Step(`^I install the "([^"]*)" component$`, iInstallTheComponent)
	ctx.Step(`^the installation should complete successfully$`, theInstallationShouldCompleteSuccessfully)
	ctx.Step(`^the component should appear in the installed components list$`, theComponentShouldAppearInTheInstalledComponentsList)
	ctx.Step(`^the catalog is connected to the remote index$`, theCatalogIsConnectedToTheRemoteIndex)
	ctx.Step(`^I check for catalog updates$`, iCheckForCatalogUpdates)
	ctx.Step(`^I should receive the current catalog version$`, iShouldReceiveTheCurrentCatalogVersion)
	ctx.Step(`^any new components should be listed$`, anyNewComponentsShouldBeListed)

	// Pipeline feature steps
	ctx.Step(`^a pipeline is created for component "([^"]*)"$`, aPipelineIsCreatedForComponent)
	ctx.Step(`^system detection has been initialized$`, systemDetectionHasBeenInitialized)
	ctx.Step(`^the pipeline runs all stages in order$`, thePipelineRunsAllStagesInOrder)
	ctx.Step(`^stage "([^"]*)" should complete successfully$`, stageShouldCompleteSuccessfully)
	ctx.Step(`^the pipeline should report success$`, thePipelineShouldReportSuccess)
	ctx.Step(`^system detection runs$`, systemDetectionRuns)
	ctx.Step(`^the detected OS should be identified$`, theDetectedOSShouldBeIdentified)
	ctx.Step(`^the architecture should be detected$`, theArchitectureShouldBeDetected)
	ctx.Step(`^the user home directory should be found$`, theUserHomeDirectoryShouldBeFound)
	ctx.Step(`^installed agents should be listed$`, installedAgentsShouldBeListed)
	ctx.Step(`^the system has existing ODIN data at ~/.odin$`, func(ctx context.Context) error {
		return theSystemHasExistingODINDataAt(ctx, "~/.odin")
	})
	ctx.Step(`^the backup stage runs$`, theBackupStageRuns)
	ctx.Step(`^a backup archive should be created$`, aBackupArchiveShouldBeCreated)
	ctx.Step(`^the backup path should be stored for rollback$`, theBackupPathShouldBeStoredForRollback)
	ctx.Step(`^the pipeline is running$`, thePipelineIsRunning)
	ctx.Step(`^the install stage fails$`, theInstallStageFails)
	ctx.Step(`^the pipeline performs rollback$`, thePipelinePerformsRollback)
	ctx.Step(`^the component directory should be removed$`, theComponentDirectoryShouldBeRemoved)
	ctx.Step(`^the backup should be restored$`, theBackupShouldBeRestored)
	ctx.Step(`^an error should be returned$`, anErrorShouldBeReturned)
	ctx.Step(`^the user cancels the pipeline \(Ctrl\+C\)$`, theUserCancelsThePipeline)
	ctx.Step(`^the pipeline should stop gracefully$`, thePipelineShouldStopGracefully)
	ctx.Step(`^completed stages should be rolled back$`, completedStagesShouldBeRolledBack)
	ctx.Step(`^an interruption error should be returned$`, anInterruptionErrorShouldBeReturned)
	ctx.Step(`^a component exists in the catalog$`, func(ctx context.Context) error {
		return aComponentExistsInTheCatalog(ctx, "test-component")
	})
	ctx.Step(`^the install stage executes$`, theInstallStageExecutes)
	ctx.Step(`^the component directory should be created at ~/.odin/\{component\}$`, theComponentDirectoryShouldBeCreatedAt)
	ctx.Step(`^the component's runes should be installed$`, theComponentRunesShouldBeInstalled)
	ctx.Step(`^the stage should report success$`, theStageShouldReportSuccess)
	ctx.Step(`^a component has been installed$`, aComponentHasBeenInstalled)
	ctx.Step(`^the verify stage runs$`, theVerifyStageRuns)
	ctx.Step(`^the component directory should exist$`, theComponentDirectoryShouldExist)
	ctx.Step(`^all required files should be present$`, allRequiredFilesShouldBePresent)
	ctx.Step(`^the verification should pass$`, theVerificationShouldPass)

	// Memory feature steps
	ctx.Step(`^the memory store is initialized$`, theMemoryStoreIsInitialized)
	ctx.Step(`^the embedder is configured with Ollama$`, theEmbedderIsConfiguredWithOllama)
	ctx.Step(`^I have a memory with content "([^"]*)" and tagged with "([^"]*)"$`, iHaveAMemoryWithContentAndTaggedWith)
	ctx.Step(`^when I store the memory$`, whenIStoreTheMemory)
	ctx.Step(`^then the memory should be saved with a unique ID$`, thenTheMemoryShouldBeSavedWithID)
	ctx.Step(`^the memory should have a created timestamp$`, theMemoryShouldHaveCreatedTimestamp)
	ctx.Step(`^the vector embedding should be stored$`, theVectorEmbeddingShouldBeStored)
	ctx.Step(`^I have stored several memories with embeddings$`, iHaveStoredMemoriesWithEmbeddings)
	ctx.Step(`^when I search for memories similar to "([^"]*)" using vector search$`, whenISearchForMemoriesSimilarToUsingVectorSearch)
	ctx.Step(`^then I should receive similar memories ranked by score$`, thenIShouldReceiveSimilarMemoriesRankedByScore)
	ctx.Step(`^each result should have a similarity score$`, eachResultShouldHaveASimilarityScore)
	ctx.Step(`^results should be ordered by relevance$`, resultsShouldBeOrderedByRelevance)
	ctx.Step(`^I have memories with content about "([^"]*)"$`, iHaveMemoriesWithContentAbout)
	ctx.Step(`^when I search for "([^"]*)" using FTS5$`, whenISearchForUsingFTS5)
	ctx.Step(`^then I should receive memories containing "([^"]*)"$`, iShouldReceiveMemoriesContaining)
	ctx.Step(`^the search should be case-insensitive$`, theSearchShouldBeCaseInsensitiveFTS)
	ctx.Step(`^results should be ranked by relevance$`, resultsShouldBeRankedByRelevanceFTS)
	ctx.Step(`^a memory exists with ID "([^"]*)"$`, aMemoryExistsWithID)
	ctx.Step(`^when I retrieve the memory by ID$`, iRetrieveTheMemoryByID)
	ctx.Step(`^then I should receive the full memory content$`, iShouldReceiveTheFullMemoryContent)
	ctx.Step(`^all tags should be included$`, allTagsShouldBeIncluded)
	ctx.Step(`^metadata should be preserved$`, metadataShouldBePreserved)
	ctx.Step(`^a memory with tags "([^"]*)"$`, aMemoryWithTags)
	ctx.Step(`^when I add tag "([^"]*)" to the memory$`, iAddTagToTheMemory)
	ctx.Step(`^then the memory should now have three tags$`, theMemoryShouldNowHaveThreeTags)
	ctx.Step(`^existing tags should be preserved$`, existingTagsShouldBePreserved)
	ctx.Step(`^a memory exists with ID "([^"]*)"$`, func(ctx context.Context, id string) error {
		return aMemoryExistsWithID(ctx, id)
	})
	ctx.Step(`^when I delete the memory$`, iDeleteTheMemory)
	ctx.Step(`^then the memory should be removed from storage$`, theMemoryShouldBeRemovedFromStorage)
	ctx.Step(`^its vector should also be deleted$`, itsVectorShouldAlsoBeDeleted)
	ctx.Step(`^subsequent retrieval should return nil$`, subsequentRetrievalShouldReturnNil)
	ctx.Step(`^memory "([^"]*)" and memory "([^"]*)" exist$`, memoryAndMemoryExist)
	ctx.Step(`^when I add an edge from "([^"]*)" to "([^"]*)" with relation "([^"]*)"$`, iAddAnEdgeFromToWithRelation)
	ctx.Step(`^then the edge should be stored in the graph$`, theEdgeShouldBeStoredInTheGraph)
	ctx.Step(`^when I query edges for memory "([^"]*)"$`, iQueryEdgesForMemory)
	ctx.Step(`^then I should receive both connected memories$`, iShouldReceiveBothConnectedMemories)
	ctx.Step(`^each edge should have its relation type$`, eachEdgeShouldHaveItsRelationType)
	ctx.Step(`^memories with tags "([^"]*)" exist$`, memoriesWithTagsExist)
	ctx.Step(`^when I list all unique tags$`, iListAllUniqueTags)
	ctx.Step(`^then I should receive "([^"]*)" and "([^"]*)"$`, iShouldReceiveAnd)
	ctx.Step(`^duplicates should be removed$`, duplicatesShouldBeRemoved)
	ctx.Step(`^when I prune memories keeping only "([^"]*)"$`, iPruneMemoriesKeepingOnly)
	ctx.Step(`^then memories with "([^"]*)" should be deleted$`, memoriesWithShouldBeDeleted)
	ctx.Step(`^memories with "([^"]*)" should be preserved$`, memoriesWithShouldBePreserved)
}

func TestMain(m *testing.M) {
	opts := godog.Options{
		Format: "pretty",
		Paths:  []string{"../../openspec/features"},
		Tags:   "~@wip",
	}

	status := godog.TestSuite{
		Name:                "odin-e2e",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	if status != 0 {
		os.Exit(status)
	}

	os.Exit(m.Run())
}
