import { FaGoogle, FaApple, FaMeta } from 'react-icons/fa6'
import { Button } from '@/components/ui/button'

export function SocialAuth() {
  const handleSocialLogin = (provider: string) => {
    console.log(`Logging in with ${provider}`)
    // In a microservices setup, this would typically redirect to your 
    // Go Auth service's OAuth endpoint: e.g., window.location.href = `/api/auth/${provider}`
  }

  return (
    <div className="space-y-4 w-full">
      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <span className="w-full border-t border-muted" />
        </div>
        <div className="relative flex justify-center text-xs uppercase">
          <span className="bg-background px-2 text-muted-foreground">
            Or continue with
          </span>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-3">
        <Button
          variant="outline"
          type="button"
          onClick={() => handleSocialLogin('google')}
          className="hover:bg-muted"
        >
          <FaGoogle className="h-4 w-4" />
          <span className="sr-only">Google</span>
        </Button>

        <Button
          variant="outline"
          type="button"
          onClick={() => handleSocialLogin('apple')}
          className="hover:bg-muted"
        >
          <FaApple className="h-4 w-4" />
          <span className="sr-only">Apple</span>
        </Button>

        <Button
          variant="outline"
          type="button"
          onClick={() => handleSocialLogin('meta')}
          className="hover:bg-muted"
        >
          <FaMeta className="h-4 w-4" />
          <span className="sr-only">Meta</span>
        </Button>
      </div>
    </div>
  )
}
