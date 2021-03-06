import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { Query } from "react-apollo";

import AppContent from "/components/AppContent";
import NotFoundView from "/components/views/NotFoundView";
import EntityDetailsContainer from "/components/partials/EntityDetailsContainer";
import Loader from "/components/Loader";

const query = gql`
  query EntityDetailsContentQuery($ns: NamespaceInput!, $name: String!) {
    entity(ns: $ns, name: $name) {
      ...EntityDetailsContainer_entity
    }
  }

  ${EntityDetailsContainer.fragments.entity}
`;

class EntityDetailsContent extends React.PureComponent {
  static propTypes = {
    match: PropTypes.object.isRequired,
  };

  render() {
    const { match } = this.props;
    const { organization, environment, ...params } = match.params;
    const ns = { organization, environment };

    return (
      <Query
        query={query}
        fetchPolicy="cache-and-network"
        variables={{ ...params, ns }}
      >
        {({ data: { entity } = {}, loading }) => {
          if (!loading && !entity) return <NotFoundView />;

          return (
            <AppContent>
              <Loader loading={loading} passthrough>
                {entity && <EntityDetailsContainer entity={entity} />}
              </Loader>
            </AppContent>
          );
        }}
      </Query>
    );
  }
}

export default EntityDetailsContent;
