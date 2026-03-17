'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { useAuthStore } from '@/store/auth'

export default function MyProfilePage() {
  const router = useRouter()
  const { account, activeProfile, logout, isSignedIn } = useAuthStore()
  const profile = activeProfile()

  useEffect(() => {
    if (!isSignedIn) {
      router.push('/auth/signin?redirect=/my-profile')
    }
  }, [isSignedIn, router])

  if (!isSignedIn || !profile) return null

  return (
    <div className="container mx-auto py-10 px-4">
      <h1 className="text-3xl font-bold mb-6">My Profile</h1>
      <div className="grid gap-6">
        <div className="p-6 border rounded-lg bg-card shadow-sm">
          <h2 className="text-xl font-semibold mb-4">Profile Information</h2>
          <div className="space-y-2">
            <Avatar>
              <AvatarImage src={profile.avatarUrl || ''} />
              <AvatarFallback>{profile.firstName[0]}</AvatarFallback>
            </Avatar>
            <p><strong>Email:</strong> {account?.email}</p>
            <p><strong>Name:</strong> {profile.firstName} {profile.lastName}</p>
          </div>
          <div className="mt-6 flex gap-3">
            <Button asChild>
              <Link href="/my-profile/edit">Edit Profile</Link>
            </Button>
            <Button variant="destructive" onClick={() => logout()}>Log out</Button>
          </div>
        </div>
      </div>
    </div>
  )
}