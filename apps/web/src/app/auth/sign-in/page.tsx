import { getCurrentUser } from '@/lib/api-client';
import { headers } from 'next/headers';
import { redirect } from 'next/navigation';

import SignInForm from './sign-in-form';

export const metadata = { title: 'Sign in' };

export default async function SignInPage() {
  // Already signed in: skip the form.
  const requestHeaders = await headers();
  const user = await getCurrentUser(requestHeaders.get('cookie') ?? undefined);
  if (user) redirect('/dashboard/overview');

  return <SignInForm />;
}
