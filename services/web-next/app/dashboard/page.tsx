'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Card, CardHeader } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/store/auth'

export default function DashboardPage() {
  const router = useRouter()
  const { account, activeProfile, isSignedIn } = useAuthStore()
  const profile = activeProfile()

  useEffect(() => {
    if (!isSignedIn) {
      router.push('/auth/signin?redirect=/dashboard')
    }
  }, [isSignedIn, router])

  if (!isSignedIn || !profile) return null

  return (
    <div className="p-6">
      <Card className="max-w-md">
        <CardHeader className="space-y-1">
          <div className="text-2xl font-bold">Dashboard</div>
          <div className="text-sm text-muted-foreground">
            Logged in as:{' '}
            <span className="font-medium">{account?.email}</span>
          </div>
        </CardHeader>
        <div className="p-6 pt-0">
          <div className="flex items-center gap-4">
            <div className="h-10 w-10 rounded-full bg-primary flex items-center justify-center text-white font-bold">
              {profile.firstName[0]}
            </div>
            <div>
              <p className="text-sm font-medium leading-none">
                {profile.firstName} {profile.lastName}
              </p>
              <p className="text-xs text-muted-foreground">Active Profile</p>
            </div>
            <Button asChild>
              <Link href="/my-account">My Account</Link>
            </Button>
          </div>
        </div>
      </Card>
    </div>
  )
}