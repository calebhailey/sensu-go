package graphql

import (
	"errors"
	"sort"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/eval"
)

var _ schema.EnvironmentFieldResolvers = (*envImpl)(nil)

//
// Implement EnvironmentFieldResolvers
//

type envImpl struct {
	orgCtrl    actions.OrganizationsController
	checksCtrl actions.CheckController
	entityCtrl actions.EntityController
	eventsCtrl eventQuerier
}

func newEnvImpl(store store.Store, getter types.QueueGetter) *envImpl {
	eventsCtrl := actions.NewEventController(store, nil)
	return &envImpl{
		orgCtrl:    actions.NewOrganizationsController(store),
		checksCtrl: actions.NewCheckController(store, getter),
		entityCtrl: actions.NewEntityController(store),
		eventsCtrl: eventsCtrl,
	}
}

// ID implements response to request for 'id' field.
func (r *envImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.EnvironmentTranslator.EncodeToString(p.Source), nil
}

// Name implements response to request for 'name' field.
func (r *envImpl) Name(p graphql.ResolveParams) (string, error) {
	env := p.Source.(*types.Environment)
	return env.Name, nil
}

// Description implements response to request for 'description' field.
func (r *envImpl) Description(p graphql.ResolveParams) (string, error) {
	env := p.Source.(*types.Environment)
	return env.Description, nil
}

// ColourID implements response to request for 'colourId' field.
// Experimental. Value is not persisted in any way at this time and is simply
// derived from the name.
func (r *envImpl) ColourID(p graphql.ResolveParams) (schema.MutedColour, error) {
	env := p.Source.(*types.Environment)
	num := env.Name[0] % 7
	logger.WithField("name", env.Name).WithField("num", num).Info("finding colour")
	switch num {
	case 0:
		return schema.MutedColours.BLUE, nil
	case 1:
		return schema.MutedColours.GRAY, nil
	case 2:
		return schema.MutedColours.GREEN, nil
	case 3:
		return schema.MutedColours.ORANGE, nil
	case 4:
		return schema.MutedColours.PINK, nil
	case 5:
		return schema.MutedColours.PURPLE, nil
	case 6:
		return schema.MutedColours.YELLOW, nil
	}
	return "", errors.New("exhausted list of colours")
}

// Organization implements response to request for 'organization' field.
func (r *envImpl) Organization(p graphql.ResolveParams) (interface{}, error) {
	env := p.Source.(*types.Environment)
	org, err := r.orgCtrl.Find(p.Context, env.Organization)
	return handleControllerResults(org, err)
}

// Checks implements response to request for 'checks' field.
func (r *envImpl) Checks(p schema.EnvironmentChecksFieldResolverParams) (interface{}, error) {
	env := p.Source.(*types.Environment)
	ctx := types.SetContextFromResource(p.Context, env)
	records, err := r.checksCtrl.Query(ctx)
	if err != nil {
		return nil, err
	}

	// apply filters
	var filteredChecks []*types.CheckConfig
	filter := p.Args.Filter
	if len(filter) > 0 {
		predicate, err := eval.NewPredicate(filter)
		if err != nil {
			logger.WithError(err).Debug("error with given predicate")
		} else {
			for _, record := range records {
				if matched, err := predicate.Eval(record); err != nil {
					logger.WithError(err).Debug("unable to filter record")
				} else if matched {
					filteredChecks = append(filteredChecks, record)
				}
			}
		}
	} else {
		filteredChecks = records
	}

	// sort records
	sort.Sort(types.SortCheckConfigsByName(
		filteredChecks,
		p.Args.OrderBy == schema.CheckListOrders.NAME,
	))

	// paginate
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredChecks))
	return newOffsetContainer(
		filteredChecks[l:h],
		len(filteredChecks),
		p.Args.Offset,
		p.Args.Limit,
	), nil
}

// Entities implements response to request for 'entities' field.
func (r *envImpl) Entities(p schema.EnvironmentEntitiesFieldResolverParams) (interface{}, error) {
	env := p.Source.(*types.Environment)
	ctx := types.SetContextFromResource(p.Context, env)
	records, err := r.entityCtrl.Query(ctx)
	if err != nil {
		return nil, err
	}

	// apply filters
	var filteredEntities []*types.Entity
	filter := p.Args.Filter
	if len(filter) > 0 {
		predicate, err := eval.NewPredicate(filter)
		if err != nil {
			logger.WithError(err).Debug("error with given predicate")
		} else {
			for _, event := range records {
				if matched, err := predicate.Eval(event); err != nil {
					logger.WithError(err).Debug("unable to filter record")
				} else if matched {
					filteredEntities = append(filteredEntities, event)
				}
			}
		}
	} else {
		filteredEntities = records
	}

	// sort records
	switch p.Args.OrderBy {
	case schema.EntityListOrders.LASTSEEN:
		sort.Sort(types.SortEntitiesByLastSeen(filteredEntities))
	default:
		sort.Sort(types.SortEntitiesByID(
			filteredEntities,
			p.Args.OrderBy == schema.EntityListOrders.ID,
		))
	}

	// paginate
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredEntities))
	return newOffsetContainer(
		filteredEntities[l:h],
		len(filteredEntities),
		p.Args.Offset,
		p.Args.Limit,
	), nil
}

// Events implements response to request for 'events' field.
func (r *envImpl) Events(p schema.EnvironmentEventsFieldResolverParams) (interface{}, error) {
	env := p.Source.(*types.Environment)
	ctx := types.SetContextFromResource(p.Context, env)
	records, err := r.eventsCtrl.Query(ctx, "", "")
	if err != nil {
		return nil, err
	}

	// apply filters
	var filteredEvents []*types.Event
	filter := p.Args.Filter
	if len(filter) > 0 {
		predicate, err := eval.NewPredicate(filter)
		if err != nil {
			logger.WithError(err).Debug("error with given predicate")
		} else {
			for _, event := range records {
				if matched, err := predicate.Eval(event); err != nil {
					logger.WithError(err).Debug("unable to filter event")
				} else if matched {
					filteredEvents = append(filteredEvents, event)
				}
			}
		}
	} else {
		filteredEvents = records
	}

	// sort records
	if p.Args.OrderBy == schema.EventsListOrders.SEVERITY {
		sort.Sort(types.EventsBySeverity(filteredEvents))
	} else {
		sort.Sort(types.EventsByTimestamp(
			filteredEvents,
			p.Args.OrderBy == schema.EventsListOrders.NEWEST,
		))
	}

	// pagination
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredEvents))
	return newOffsetContainer(
		filteredEvents[l:h],
		len(filteredEvents),
		p.Args.Offset,
		p.Args.Limit,
	), nil
}

// CheckHistory implements response to request for 'checkHistory' field.
func (r *envImpl) CheckHistory(p schema.EnvironmentCheckHistoryFieldResolverParams) (interface{}, error) {
	env := p.Source.(*types.Environment)
	ctx := types.SetContextFromResource(p.Context, env)
	records, err := r.eventsCtrl.Query(ctx, "", "")
	if err != nil {
		return []types.CheckHistory{}, err
	}

	// Accumulate history
	history := []types.CheckHistory{}
	for _, record := range records {
		if record.Check == nil {
			continue
		}
		latest := types.CheckHistory{
			Executed: record.Check.Executed,
			Status:   record.Check.Status,
		}
		history = append(history, latest)
		history = append(history, record.Check.History...)
	}

	// Sort
	sort.Sort(types.ByExecuted(history))

	// Limit
	limit := clampInt(p.Args.Limit, 0, len(history))
	return history[0:limit], nil
}
