package app

import (
	"net/http"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-party-mgmt-svc/internal/application/command"
	"github.com/medhen/pc-party-mgmt-svc/internal/application/query"
	"github.com/medhen/pc-party-mgmt-svc/internal/infrastructure/elasticsearch"
	"github.com/medhen/pc-party-mgmt-svc/internal/infrastructure/fayda"
	"github.com/medhen/pc-party-mgmt-svc/internal/infrastructure/postgres"
	"github.com/medhen/pc-party-mgmt-svc/internal/api/rest"
)

// NewTestHandler returns an http.Handler wired with all real internal components for integration testing.
func NewTestHandler(dbPool *pgxpool.Pool, esClient *es.Client) http.Handler {
	uow := postgres.NewUnitOfWork(dbPool)
	searchRepo := elasticsearch.NewSearchRepository(esClient)
	faydaClient := fayda.NewClient("http://localhost:8089") // Mock URL for test handler
	
	registerCmd := command.NewRegisterPartyHandler(uow, searchRepo, faydaClient)
	addAddrCmd := command.NewAddAddressHandler(uow)
	query360 := query.NewCustomer360QueryService(dbPool)
	
	r := chi.NewRouter()
	updateConsentCmd := command.NewUpdateConsentHandler(uow)
	anonymizeCmd := command.NewAnonymizePartyHandler(uow)

	handler := rest.NewPartyHandler(registerCmd, addAddrCmd, updateConsentCmd, anonymizeCmd, query360)
	handler.RegisterRoutes(r)
	
	return r
}
