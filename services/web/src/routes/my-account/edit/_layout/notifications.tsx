import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/my-account/edit/_layout/notifications')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/my-account/edit/_layout/notifications"!</div>
}
