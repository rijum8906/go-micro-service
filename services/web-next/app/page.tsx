import Link from 'next/link'

export default async function Home() {
  const apiUrl = process.env.API_BASE_URL || 'not found'

  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <h1 className="text-2xl font-bold">Home</h1>
      <Link href="/auth/signin">Signin</Link>
      <h1>{apiUrl}</h1>
    </div>
  )
}