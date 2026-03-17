'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuthStore } from '@/store/auth'

export default function MyAccountPage() {
  const router = useRouter()
  const { account, logout, activeProfile, isSignedIn } = useAuthStore()
  const profile = activeProfile()

  useEffect(() => {
    if (!isSignedIn) {
      router.push('/auth/signin?redirect=/my-account')
    }
  }, [isSignedIn, router])

  if (!isSignedIn || !profile) return null

  return (
    <div className="container mx-auto max-w-4xl py-10 px-4">
      <h1 className="mb-8 text-3xl font-bold">My Account</h1>
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Profile</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-4">
              <Avatar className="h-16 w-16">
                <AvatarImage src={profile.avatarUrl ?? ''} />
                <AvatarFallback>
                  {profile.firstName[0]}
                  {profile.lastName?.[0]}
                </AvatarFallback>
              </Avatar>
              <div>
                <p className="text-lg font-semibold">
                  {profile.firstName} {profile.lastName}
                </p>
                <p className="text-sm text-muted-foreground">Active Profile</p>
              </div>
            </div>
            <div className="flex flex-wrap gap-3 pt-2">
              <Button asChild size="sm">
                <Link href="/my-profile/edit">Edit Profile</Link>
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Account</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm text-muted-foreground">Email</p>
              <p className="font-medium">{account?.email}</p>
            </div>
            <div className="flex flex-wrap gap-3 pt-2">
              <Button asChild size="sm" variant="outline">
                <Link href="/my-account/edit/security">Security</Link>
              </Button>
              <Button size="sm" variant="destructive" onClick={logout}>
                Log out
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}