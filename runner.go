package endly

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/viant/toolbox"
	"strings"
	"time"
)

var reportingEventSleep = 250 * time.Millisecond

type RunnerReportingFilter struct {
	Stdin                   bool
	Stdout                  bool
	Transfer                bool
	Task                    bool
	UseCase                 bool
	Action                  bool
	Deployment              bool
	SQLScript               bool
	PopulateDatastore       bool
	Sequence                bool
	RegisterDatastore       bool
	Assert                  bool
	DataMapping             bool
	HttpTrip                bool
	Workflow                bool
	WorkflowParams          bool
	OnFailureFilter         *RunnerReportingFilter
	FirstUseCaseFailureOnly bool
}

type RunnerReportingOption struct {
	Filter *RunnerReportingFilter
}

type UseCase struct {
	Description string
	ActionGroup string
	Syspath     string
	Events      []*Event
	PassedCount int
	FailedCount int
}

func (c *UseCase) AddEvent(event *Event) {
	if len(c.Events) == 0 {
		c.Events = make([]*Event, 0)
	}
	c.Events = append(c.Events, event)
}

type CliRunner struct {
	manager    Manager
	useCases   []*UseCase
	ErrorEvent *Event
}

func (r *CliRunner) AddUseCase(useCases *UseCase) {
	r.useCases = append(r.useCases, useCases)
}

func (r *CliRunner) CurrentUseCase() *UseCase {
	if len(r.useCases) == 0 {
		useCase := &UseCase{}
		r.AddUseCase(useCase)
	}
	return r.useCases[len(r.useCases)-1]
}

func (r *CliRunner) hasActiveSession(context *Context, sessionId string) bool {
	service, err := context.Service(WorkflowServiceId)
	if err != nil {
		return false
	}
	var state = service.State()
	service.Mutex().RLock()
	defer service.Mutex().RUnlock()
	return state.Has(sessionId)
}

var reportDataLayout = toolbox.DateFormatToLayout("hh:mm:ss.SSS")

func formatInput(tag, input string, event *Event) string {
	sessionInfo := aurora.Gray(fmt.Sprintf("[%-17v]", tag)).Bold()
	stdin := aurora.Blue(fmt.Sprintf(" %v", input))
	return fmt.Sprintf("%v%v", sessionInfo, stdin)
}

func formatOutput(tag, input string, event *Event) string {
	sessionInfo := aurora.Gray(fmt.Sprintf("[%-18v]", tag)).Bold()
	stdin := aurora.Green(fmt.Sprintf(" %v", input))
	return fmt.Sprintf("%v%v", sessionInfo, stdin)
}

var targetedLineLength = 80
var tagLength = 16
var tagFormat = "[%-" + toolbox.AsString(tagLength) + "v"

func formatStageEvent(name, stage, argument string, event *Event) string {
	eventType := aurora.Brown(fmt.Sprintf(tagFormat, name))
	stageInfo := aurora.Gray(fmt.Sprintf(" %15v", stage))
	argumentInfo := aurora.Brown(fmt.Sprintf("%40v]", argument))
	return fmt.Sprintf("%v%v%v", eventType, stageInfo, argumentInfo)
}

func formatError(error interface{}, event *Event) string {
	errorTag := aurora.Red("[error]").Bold()
	errorInfo := aurora.Red(fmt.Sprintf(" %v", error))
	return fmt.Sprintf("%v%v", errorTag, errorInfo)
}

func printGenericColoredEvent(eventType, name, argument string, event *Event, nameLength, argumentLength int, nameColor, argumentColor func(arg interface{}) aurora.Value) {
	eventType = fmt.Sprintf(tagFormat, eventType)
	formattedEventType := aurora.Brown(eventType)
	name = fmt.Sprintf("%"+toolbox.AsString(nameLength)+"v", name)
	argument = fmt.Sprintf("%"+toolbox.AsString(argumentLength)+"v]", argument)
	if len(eventType)+len(argument)+len(name) > targetedLineLength {
		var overflow = (len(eventType) + len(argument) + len(name)) - targetedLineLength + 1
		name = strings.Replace(name, " ", "", overflow)
	}
	nameInfo := nameColor(name)
	argumentInfo := argumentColor(argument)
	fmt.Printf("%v%v%v\n", formattedEventType, nameInfo, argumentInfo)
}
func printGenericEvent(eventType, name, argument string, event *Event, nameLength, argumentLength int) {
	printGenericColoredEvent(eventType, name, argument, event, nameLength, argumentLength, aurora.Brown, aurora.Bold)
}

func formatStartEvent(name, argument string, event *Event) string {
	var stage = fmt.Sprintf("started:%v", event.Timestamp.Format(reportDataLayout))
	return formatStageEvent(name, stage, argument, event)
}

func formatEndEvent(name, argument string, event *Event) string {
	var elapsed = fmt.Sprintf("%9.3f ", float64(event.TimeTakenMs)/1000)
	var stage = fmt.Sprintf("elapsed:%vs.", elapsed)
	return formatStageEvent(name, stage, argument, event)
}

func (r *CliRunner) reportStdin(event *Event) {
	if stdin, ok := event.Value["stdin"]; ok {
		var session = event.Value["session"]
		formattedText := formatInput(fmt.Sprintf("%-13v%7v", session, "stdin"), toolbox.AsString(stdin), event)
		fmt.Printf("%v\n", formattedText)
	}
}

func (r *CliRunner) reportStdout(event *Event) {
	if stdout, ok := event.Value["stdout"]; ok {
		var session = event.Value["session"]
		formattedText := formatInput(fmt.Sprintf("%-13v%7v", session, "stdout"), toolbox.AsString(stdout), event)
		fmt.Printf("%v\n", formattedText)
	}
}

func (r *CliRunner) reportError(event *Event) {
	if error, ok := event.Value["error"]; ok {
		formattedText := formatError(error, event)
		fmt.Printf("%v\n", formattedText)
	}
}

func (r *CliRunner) reportWorkflowStart(event *Event) {
	if request, ok := event.Value["request"]; ok {
		if runRequest, ok := request.(*WorkflowRunRequest); ok {
			var task = "*"
			if runRequest.Tasks != "" {
				task = runRequest.Tasks
			}
			var formattedText = formatStartEvent("Workflow", fmt.Sprintf("%v:%v", runRequest.Name, task), event)
			fmt.Printf("%v\n", formattedText)
		}
	}
}

func (r *CliRunner) reportWorkflowEnd(event *Event) {
	startEvent := event.StartEvent
	if request, ok := startEvent.Value["request"]; ok {
		if runRequest, ok := request.(*WorkflowRunRequest); ok {
			var formattedText = formatEndEvent("Workflow", runRequest.Name, event)
			fmt.Printf("%v\n", formattedText)
		}
	}
}

func (r *CliRunner) reportTaskStart(event *Event) {
	if taskName, ok := event.Value["name"]; ok {
		var formattedText = formatStartEvent("Workflow Task", toolbox.AsString(taskName), event)
		fmt.Printf("%v\n", formattedText)
	}
}

func (r *CliRunner) reportTaskEnd(event *Event) {
	startEvent := event.StartEvent
	if taskName, ok := startEvent.Value["name"]; ok {
		var formattedText = formatEndEvent("Workflow Task", toolbox.AsString(taskName), event)
		fmt.Printf("%v\n", formattedText)
	}
}

func (r *CliRunner) reportWorkflowActionStart(event *Event) {
	if action, ok := event.Value["action"]; ok {
		service := event.Value["service"]
		var formattedText = formatStartEvent("Workflow Action", fmt.Sprintf("%v.%v", service, action), event)
		fmt.Printf("%v\n", formattedText)
	}
}

func (r *CliRunner) reportWorkflowActionEnd(event *Event) {
	startEvent := event.StartEvent
	if action, ok := startEvent.Value["action"]; ok {
		service := startEvent.Value["service"]
		var formattedText = formatEndEvent("Workflow Action", fmt.Sprintf("%v.%v", service, action), event)
		fmt.Printf("%v\n", formattedText)
	}
}

func (r *CliRunner) reportTransfer(event *Event) {
	expand, _ := event.Value["expand"]
	var expandInfo = fmt.Sprintf("expand:%v", expand)
	var formattedEvent = formatStartEvent("Copy", expandInfo, event)
	fmt.Printf("%v\n", formattedEvent)

	formattedSource := formatInput("SourceURL", toolbox.AsString(event.Value["source"]), event)
	fmt.Printf("%v\n", formattedSource)

	formattedTarget := formatOutput("TargetURL", toolbox.AsString(event.Value["target"]), event)
	fmt.Printf("%v\n", formattedTarget)
}

func (r *CliRunner) reportDeploymentStart(event *Event) {
	if request, ok := event.Value["request"]; ok {

		if deploymentRequest, ok := request.(*DeploymentDeployRequest); ok {
			var formattedText = formatStartEvent(fmt.Sprintf("Deploy %v", deploymentRequest.AppName), fmt.Sprintf("sdk:%v:%v, force:%v", deploymentRequest.Sdk, deploymentRequest.SdkVersion, deploymentRequest.Force), event)
			fmt.Printf("%v\n", formattedText)
		}
	}
}

func (r *CliRunner) reportDeploymentEnd(event *Event) {
	var startEvent = event.StartEvent
	if request, ok := startEvent.Value["request"]; ok {
		if deploymentRequest, ok := request.(*DeploymentDeployRequest); ok {
			var formattedText = formatEndEvent(fmt.Sprintf("Deploy %v", deploymentRequest.AppName), fmt.Sprintf("sdk:%v:%v, force:%v", deploymentRequest.Sdk, deploymentRequest.SdkVersion, deploymentRequest.Force), event)
			fmt.Printf("%v\n", formattedText)
		}
	}
}

func (r *CliRunner) reportDsUnitRegister(event *Event) {
	if value, ok := event.Value["request"]; ok {
		if request, ok := value.(*DsUnitRegisterRequest); ok {
			printGenericEvent(fmt.Sprintf("Datastore %v", request.Datastore), "", fmt.Sprintf("%v:%v", request.Config.DriverName, request.Config.Descriptor), event, 1, 59)

		}
	}
}

func (runner *CliRunner) reportDsUnitMapping(event *Event) {
	if value, ok := event.Value["request"]; ok {
		if request, ok := value.(*DsUnitMappingRequest); ok {
			for _, mapping := range request.Mappings {
				printGenericEvent("Mapping", mapping.Name, mapping.URL, event, 10, 20)
			}
		}
	}
}

func (runner *CliRunner) reportDsUnitSequence(event *Event) {
	if value, ok := event.Value["response"]; ok {
		if serviceResponse, ok := value.(*ServiceResponse); ok {
			if response, ok := serviceResponse.Response.(*DsUnitTableSequenceResponse); ok {
				for k, v := range response.Sequences {
					printGenericEvent("Sequence", k, toolbox.AsString(v), event, 46, 15)
				}
			}
		}
	}
}

func (r *CliRunner) reportPopulateDatestore(event *Event) {
	if datastore, ok := event.Value["datastore"]; ok {
		printGenericEvent(fmt.Sprintf("Populate  %v", datastore), fmt.Sprintf("%v", event.Value["table"]), fmt.Sprintf("%v rows", event.Value["rows"]), event, 46, 15)

	}
}

func (r *CliRunner) reportSQLScript(event *Event) {
	if datastore, ok := event.Value["datastore"]; ok {
		printGenericEvent(fmt.Sprintf("SQLScript %v", datastore), "", fmt.Sprintf("%v", event.Value["url"]), event, 1, 59)
	}
	//	s.AddEvent(context, "SQLScript", Pairs("datasore", request.Datastore, "url", script.URL), Info)
}

func (r *CliRunner) reportSleep(event *Event) {
	if value, ok := event.Value["sleepTime"]; ok {
		var formattedText = formatStartEvent("Sleep", fmt.Sprintf("%v ms", value), event)
		fmt.Printf("%v\n", formattedText)
	}
}

func (r *CliRunner) reportActionGroup(event *Event, filter *RunnerReportingFilter) {
	if name, ok := event.Value["name"]; ok {

		//remove this use vcase from previous use case
		previousUseCase := r.CurrentUseCase()
		previousUseCase.Events = previousUseCase.Events[:len(previousUseCase.Events)-1]

		useCase := &UseCase{
			Description: fmt.Sprintf(" %v", event.Value["description"]),
			ActionGroup: fmt.Sprintf("%v", name),
			Syspath:     fmt.Sprintf("%v ", event.Value["subPath"]),
		}
		useCase.AddEvent(event)
		r.AddUseCase(useCase)
		if filter.UseCase {
			printGenericEvent(fmt.Sprintf("UseCase %v ", useCase.ActionGroup), useCase.Syspath, useCase.Description, event, 10, 49)
		}
	}
}

func asJsonText(source interface{}) string {
	if source == nil {
		return ""
	}
	var buf = new(bytes.Buffer)
	toolbox.NewJSONEncoderFactory().Create(buf).Encode(source)
	return buf.String()
}

func (r *CliRunner) reportHttpRequestStart(event *Event) {
	if request, ok := event.Value["request"]; ok {
		if httpRequest, ok := request.(*HttpRequest); ok {
			printGenericEvent("HttpRequest ", httpRequest.Method, httpRequest.URL, event, 10, 49)
			if len(httpRequest.Header) > 0 {
				var formattedBody = formatInput("headers", asJsonText(httpRequest.Header), event)
				fmt.Printf("%v\n", formattedBody)
			}
			var formattedBody = formatInput("body", httpRequest.Body, event)
			fmt.Printf("%v\n", formattedBody)

		}
	}
}

func (r *CliRunner) reportHttpRequestEnd(event *Event) {

	if response, ok := event.Value["response"]; ok {
		if httpResponse, ok := response.(*HttpResponse); ok {

			printGenericEvent("HttpResponse ", fmt.Sprintf("%v", httpResponse.Code), "", event, 10, 49)
			if len(httpResponse.Header) > 0 {
				var formattedBody = formatOutput("headers", asJsonText(httpResponse.Header), event)
				fmt.Printf("%v\n", formattedBody)
			}
			var formattedBody = formatOutput("body", httpResponse.Body, event)
			fmt.Printf("%v\n", formattedBody)

		}
	}
}

func (r *CliRunner) reportValidatorStart(event *Event) {
	if request, ok := event.Value["request"]; ok {
		if assertRequest, ok := request.(*ValidatorAssertRequest); ok {
			var expected = assertRequest.Expected
			if toolbox.IsSlice(expected) || toolbox.IsMap(expected) {
				expected = asJsonText(expected)
				expected = strings.Trim(toolbox.AsString(expected), " \n")
			}
			var actual = assertRequest.Actual
			if toolbox.IsSlice(actual) || toolbox.IsMap(actual) {
				actual = asJsonText(actual)
				actual = strings.Trim(toolbox.AsString(actual), " \n")
			}
			var formattedExpected = formatInput("Assert expected", fmt.Sprintf("%v", expected), event)
			fmt.Printf("%v\n", formattedExpected)

			var formattedActual = formatOutput("Assert actual", fmt.Sprintf("%v", actual), event)
			fmt.Printf("%v\n", formattedActual)
		}

	}

}

func (r *CliRunner) reportAssertionInfo(event *Event, filter *RunnerReportingFilter) {
	useCase := r.CurrentUseCase()
	if serviceResponse, ok := event.Value["response"]; ok {
		if response, ok := serviceResponse.(*ServiceResponse); ok {

			if assertionInfo, ok := response.Response.(*ValidatorAssertionInfo); ok {
				useCase.PassedCount += assertionInfo.TestPassed
				useCase.FailedCount += len(assertionInfo.TestFailed)

				if filter.Assert {
					printGenericColoredEvent("Assertion", "Passed", fmt.Sprintf("%v", assertionInfo.TestPassed), event, 20, 59, aurora.Green, aurora.Bold)
					printGenericColoredEvent("Assertion", "Failed", fmt.Sprintf("%v", len(assertionInfo.TestFailed)), event, 20, 59, aurora.Red, aurora.Bold)

					for _, failed := range assertionInfo.TestFailed {
						printGenericColoredEvent("Failure", "", failed, event, 1, 69, aurora.Red, aurora.Gray)
					}
				}
			}

		}

	}
}

func (r *CliRunner) reportBuild(event *Event, filter *RunnerReportingFilter) {
	if serviceResponse, ok := event.Value["request"]; ok {
		if buildRequest, ok := serviceResponse.(*BuildRequest); ok {
			printGenericEvent("Build ", fmt.Sprintf("%v", buildRequest.BuildSpec.Name), buildRequest.Target.URL, event, 10, 51)
		}
	}

}

func (r *CliRunner) reportCheckout(event *Event, filter *RunnerReportingFilter) {
	if serviceResponse, ok := event.Value["request"]; ok {
		if checkoutRequest, ok := serviceResponse.(*VcCheckoutRequest); ok {
			printGenericEvent("Checkout ", fmt.Sprintf("%v", checkoutRequest.Origin.URL), checkoutRequest.Target.URL, event, 30, 31)
		}
	}
}

func (r *CliRunner) reportEvent(context *Context, event *Event, filter *RunnerReportingFilter) error {
	useCase := r.CurrentUseCase()
	useCase.AddEvent(event)

	if event.Level > Debug {
		return nil
	}

	if strings.HasPrefix(event.Type, "ManagedCommandRequest") {
		return nil
	}

	switch event.Type {

	case "WorkflowRunRequest.Start":
		if !filter.Workflow {
			return nil
		}
		r.reportWorkflowStart(event)
	case "WorkflowRunRequest.End":
		if !filter.Workflow {
			return nil
		}
		r.reportWorkflowEnd(event)

	case "ServiceAction.Start":
		if !filter.Action {
			return nil
		}
		r.reportWorkflowActionStart(event)
	case "ServiceAction.End":
		if !filter.Action {
			return nil
		}
		r.reportWorkflowActionEnd(event)
	case "WorkflowTask.Start":
		if !filter.Task {
			return nil
		}
		r.reportTaskStart(event)
	case "WorkflowTask.End":
		if !filter.Task {
			return nil
		}
		r.reportTaskEnd(event)

	case "Execution.Start":
		if !filter.Stdin {
			return nil
		}
		r.reportStdin(event)
	case "Execution.End":
		if !filter.Stdout {
			return nil
		}
		r.reportStdout(event)

	case "Transfer.Start":
		if !filter.Transfer {
			return nil
		}
		r.reportTransfer(event)
	case "Transfer.End":

	case "DeploymentDeployRequest.Start":
		if !filter.Deployment {
			return nil
		}
		r.reportDeploymentStart(event)

	case "DeploymentDeployRequest.End":
		if !filter.Deployment {
			return nil
		}
		r.reportDeploymentEnd(event)

	case "DsUnitRegisterRequest.Start":
		if !filter.RegisterDatastore {
			return nil
		}
		r.reportDsUnitRegister(event)
	case "DsUnitMappingRequest.Start":
		if !filter.DataMapping {
			return nil
		}
		r.reportDsUnitMapping(event)
	case "DsUnitTableSequenceRequest.End":
		if !filter.Sequence {
			return nil
		}
		r.reportDsUnitSequence(event)
	case "PopulateDatastore":
		r.reportPopulateDatestore(event)
		if !filter.PopulateDatastore {
			return nil
		}

	case "SQLScript":
		if !filter.SQLScript {
			return nil
		}
		r.reportSQLScript(event)

	case "ActionGroup":
		r.reportActionGroup(event, filter)
	case "HttpRequest.Start":
		if !filter.HttpTrip {
			return nil
		}
		r.reportHttpRequestStart(event)

	case "HttpRequest.End":
		if !filter.HttpTrip {
			return nil
		}
		r.reportHttpRequestEnd(event)

	case "Error":
		r.reportError(event)
		r.ErrorEvent = event
	case "Sleep":
		r.reportSleep(event)
	case "ValidatorAssertRequest.Start":
		if !filter.Assert {
			return nil
		}
		r.reportValidatorStart(event)
	case "ValidatorAssertRequest.End":
		r.reportAssertionInfo(event, filter)

	case "VcCheckoutRequest.Start":
		r.reportCheckout(event, filter)
	case "BuildRequest.Start":
		r.reportBuild(event, filter)

	case "ManagedCommandRequest.Start", "ManagedCommandRequest.End",
		"ServiceStatusRequest.Start", "ServiceStatusRequest.End",
		"DockerRunRequest.Start", "DockerRunRequest.End",
		"SdkSetRequest.Start", "SdkSetRequest.End",
		"TransferCopyRequest.Start", "TransferCopyRequest.End",
		"OpenSession.Start", "OpenSession.End",
		"CloseSession.Start", "CloseSession.End",
		"DsUnitRegisterRequest.End",
		"VcCheckoutRequest.End",
		"CommandRequest.Start", "CommandRequest.End",
		"DsUnitMappingRequest.End",
		"DsUnitPrepareRequest.Start",
		"DsUnitPrepareRequest.End",
		"DsUnitTableSequenceRequest.Start",
		"SendHttpRequest.Start", "SendHttpRequest.End",
		"Nop.Start", "Nop.End",
		"ProcessStopRequest.Start", "ProcessStopRequest.End",
		"ProcessStatusRequest.Start", "ProcessStatusRequest.End":
		//ignore

	default:
		fmt.Printf("[%v]%v\n", event.Type, event.Info())

	}

	return nil

}

func (r *CliRunner) reportEvents(context *Context, sessionId string, filter *RunnerReportingFilter) error {
	service, err := context.Service(EventReporterServiceId)
	if err != nil {
		return err
	}

	time.Sleep(time.Second)
	var firstEvent *Event
	var lastEvent *Event

	if context.Workflow() != nil {
		fmt.Printf("%v\n", aurora.Bold(fmt.Sprintf("[Started: %68v]", context.Workflow().Name)))
	}
	for {
		response := service.Run(context, &EventReporterRequest{
			SessionId: sessionId,
		})

		if response.Error != "" {
			return errors.New(response.Error)
		}

		reporterResponse, ok := response.Response.(*EventReporterResponse)
		if !ok {
			return fmt.Errorf("Failed to check event - unexpected reponse type: %T", response.Response)
		}
		if len(reporterResponse.Events) == 0 {
			if !r.hasActiveSession(context, sessionId) {
				break
			}
			time.Sleep(reportingEventSleep)
			continue
		}

		for _, event := range reporterResponse.Events {
			if firstEvent == nil {
				firstEvent = event
			}
			lastEvent = event
			err = r.reportEvent(context, event, filter)
			if err != nil {
				return err
			}
		}

	}

	var totalUseCaseFailed = 0
	var totalUseCasePassed = 0
	for _, useCase := range r.useCases {
		if useCase.FailedCount > 0 {
			totalUseCaseFailed++
		} else if useCase.PassedCount > 0 {
			totalUseCasePassed++
		}
	}

	fmt.Printf("totalUseCasePassed: %v %v\n", totalUseCasePassed, totalUseCaseFailed)
	if totalUseCasePassed > 0 || totalUseCaseFailed > 0 {
		printGenericColoredEvent("Summary", "UseCases Passed", toolbox.AsString(totalUseCasePassed), nil, 20, 51, aurora.Green, aurora.Bold)
		printGenericColoredEvent("Summary", "UseCases Failed", toolbox.AsString(totalUseCaseFailed), nil, 20, 51, aurora.Red, aurora.Bold)
	}

	if totalUseCaseFailed > 0 && filter.OnFailureFilter != nil {
		for _, useCase := range r.useCases {
			if useCase.FailedCount > 0 {
				for _, event := range useCase.Events {
					err = r.reportEvent(context, event, filter.OnFailureFilter)
					if err != nil {
						return err
					}
				}
				if filter.FirstUseCaseFailureOnly {
					break
				}

			}
		}
	}
	r.reportSummary(firstEvent, lastEvent, totalUseCaseFailed)
	return nil
}

func (r *CliRunner) reportSummary(firstEvent *Event, lastEvent *Event, totalUseCaseFailed int) {
	for _, useCase := range r.useCases {
		if useCase.FailedCount > 0 {
			fmt.Printf("%v\n", aurora.Red(fmt.Sprintf("[%-6v %13v: %59v]", useCase.ActionGroup, useCase.Syspath, "Failed")))
		}
	}
	if firstEvent != nil {
		var timeTaken = lastEvent.Timestamp.UnixNano() - firstEvent.Timestamp.UnixNano()
		var elapsed = fmt.Sprintf("%9.3f ", float64(timeTaken)/float64(time.Millisecond)/1000)
		fmt.Printf("%v\n", aurora.Bold(fmt.Sprintf("[Elapsed: %70vs.]", elapsed)))
	}
	if totalUseCaseFailed > 0 || r.ErrorEvent != nil {
		fmt.Printf("%v\n", aurora.Red(fmt.Sprintf("[Status: %73v]", "ERROR")))
	} else {
		fmt.Printf("%v\n", aurora.Green(fmt.Sprintf("[Status: %73v]", "SUCCESS")))
	}
}

func (r *CliRunner) Run(workflowRunRequestURL string) error {
	request := &WorkflowRunRequest{}
	resource := NewResource(workflowRunRequestURL)
	err := resource.JsonDecode(request)
	if err != nil {
		return err
	}
	context := r.manager.NewContext(toolbox.NewContext())
	service, err := context.Service(WorkflowServiceId)
	if err != nil {
		return err
	}
	runnerOption := &RunnerReportingOption{}
	err = resource.JsonDecode(runnerOption)
	if err != nil {
		return err
	}
	fmt.Printf("RUN %v %v\n", request.Name, request.Tasks)

	request.Async = true
	response := service.Run(context, request)
	if response.Error != "" {
		return errors.New(response.Error)
	}
	workflowResponse, ok := response.Response.(*WorkflowRunResponse)
	if !ok {
		return fmt.Errorf("Failed to run workflow: %v invalid response type %T", workflowRunRequestURL, response.Response)
	}
	return r.reportEvents(context, workflowResponse.SessionId, runnerOption.Filter)
}

func NewCliRunner() *CliRunner {
	return &CliRunner{
		manager:  NewManager(),
		useCases: make([]*UseCase, 0),
	}
}
