import { Event } from 'api/events.type'
import React from 'react'
import ReactJson from 'react-json-view'

interface EventDetailProps {
  event: Event
}

const EventDetail: React.FC<EventDetailProps> = ({ event }) => (
  <ReactJson src={event} collapsed={1} displayObjectSize={false} />
)

export default EventDetail
