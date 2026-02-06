import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/my-account/')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/my-account/"!</div>
}
