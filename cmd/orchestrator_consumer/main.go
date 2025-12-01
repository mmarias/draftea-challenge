package orchestrator_consumer

import (
	orchestrator "github.com/mmarias/golearn/internal/app/orchestrator/v1"
	"github.com/mmarias/golearn/internal/entrypoint/eventbus"
	infraEventbus "github.com/mmarias/golearn/internal/infraestructure/eventbus"
	"github.com/mmarias/golearn/internal/infraestructure/publisher"
)

func Setup(bus infraEventbus.Client) {
	pub := publisher.New(bus)

	holdFundsCmd := orchestrator.NewHoldFundsCommand(pub)
	releaseFundsCmd := orchestrator.NewReleaseFundsCommand(pub)
	debitFundsCmd := orchestrator.NewDebitFundsCommand(pub)
	authorizeCmd := orchestrator.NewAuthorizeGatewayCommand(pub)
	updateStatusCmd := orchestrator.NewUpdatePaymentStatusCommand(pub)
	notifyUserCmd := orchestrator.NewNotifyUserCommand(pub)

	sagaHandler := eventbus.NewPaymentSagaHandler(
		holdFundsCmd,
		releaseFundsCmd,
		debitFundsCmd,
		authorizeCmd,
		updateStatusCmd,
		notifyUserCmd,
	)

	eventbus.SetupSagaDispatcher(bus, sagaHandler)
}
