import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import Card, { CardContent } from "material-ui/Card";
import Typography from "material-ui/Typography";
import List from "material-ui/List";
import RelativeDate from "/components/RelativeDate";
import StatusIcon from "/components/CheckStatusIcon";
import InlineLink from "/components/InlineLink";
import ListItem, {
  ListItemTitle,
  ListItemSubtitle,
} from "/components/DetailedListItem";

class EntityDetailsEvents extends React.PureComponent {
  static propTypes = {
    events: PropTypes.arrayOf(PropTypes.object).isRequired,
  };

  static fragments = {
    event: gql`
      fragment EntityDetailsEvents_event on Event {
        ns: namespace {
          org: organization
          env: environment
        }
        check {
          name
          status
        }
        entity {
          name
        }

        id
        timestamp
      }
    `,
  };

  _renderItem = eventProp => {
    const { check, entity, ns, ...event } = eventProp;

    if (check === null) {
      return null;
    }

    return (
      <ListItem key={event.id}>
        <ListItemTitle inset>
          <Typography
            component="span"
            style={{ position: "absolute", left: 0 }}
          >
            <StatusIcon statusCode={check.status} inline mutedOK small />
          </Typography>
          <InlineLink
            to={`/${ns.org}/${ns.env}/events/${entity.name}/${check.name}`}
          >
            {check.name}
          </InlineLink>
        </ListItemTitle>
        <ListItemSubtitle inset>
          Exited with status {check.status};{" "}
          <RelativeDate dateTime={event.timestamp} />
        </ListItemSubtitle>
      </ListItem>
    );
  };

  _renderItems = () => {
    const { events } = this.props;
    if (events.length === 0) {
      return <Typography>No events found.</Typography>;
    }
    return events.map(this._renderItem);
  };

  render() {
    return (
      <Card>
        <CardContent>
          <Typography variant="headline" paragraph>
            Events
          </Typography>
          <List disablePadding>{this._renderItems()}</List>
        </CardContent>
      </Card>
    );
  }
}

export default EntityDetailsEvents;
