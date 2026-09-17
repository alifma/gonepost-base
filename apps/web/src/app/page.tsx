import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import { ThemeModeToggle } from '@/components/themes/theme-mode-toggle';
import { ScrollReveal } from '@/components/scroll-reveal';
import Link from 'next/link';

export const metadata = {
  title: 'Gonepost | Operations, with clarity',
  description: 'A focused workspace for people, permissions, and operational audit trails.'
};

const capabilities = [
  { icon: 'teams' as const, title: 'People in one place', description: 'Keep your user directory current with searchable, paginated records.' },
  { icon: 'lock' as const, title: 'Access with intent', description: 'Create roles and tune permissions without scattering policy across tools.' },
  { icon: 'clock' as const, title: 'A trail you can trust', description: 'See the actions that matter, recorded by the API as they happen.' }
];

export default async function Page() {
  return (
    <main className='bg-background min-h-svh overflow-hidden'>
      <nav className='mx-auto flex w-full max-w-7xl items-center justify-between px-6 py-6 lg:px-10'>
        <Link href='/' className='flex items-center gap-3' aria-label='Gonepost home'>
          <span className='bg-foreground text-background flex size-9 items-center justify-center rounded-xl'>
            <Icons.logo className='size-4' />
          </span>
          <span className='text-sm font-semibold tracking-[0.2em] uppercase'>Gonepost</span>
        </Link>
        <div className='flex items-center gap-2'>
          <span className='text-muted-foreground hidden text-xs sm:block'>Local operations workspace</span>
          <ThemeModeToggle />
          <Button variant='outline' size='sm' render={<Link href='/auth/sign-in' />}>Sign in</Button>
        </div>
      </nav>

      <section className='mx-auto grid w-full max-w-7xl gap-14 px-6 pt-14 pb-20 lg:grid-cols-[0.9fr_1.1fr] lg:items-center lg:px-10 lg:pt-20'>
        <div className='max-w-xl'>
          <div className='text-muted-foreground mb-6 flex items-center gap-2 text-xs font-semibold tracking-[0.2em] uppercase'>
            <span className='bg-emerald-500 size-2 rounded-full' />
            A calmer operations layer
          </div>
          <h1 className='max-w-lg text-5xl leading-[0.98] font-semibold tracking-tight sm:text-6xl lg:text-7xl'>
            The work behind the work, <span className='text-muted-foreground'>made visible.</span>
          </h1>
          <p className='text-muted-foreground mt-7 max-w-md text-base leading-7 sm:text-lg'>
            Gonepost brings users, roles, permissions, and audit trails into one focused workspace for teams that need to move carefully and quickly.
          </p>
          <div className='mt-9 flex flex-wrap items-center gap-3'>
            <Button size='lg' render={<Link href='/auth/sign-in' />}>Open workspace <Icons.arrowRight /></Button>
            <Link href='#capabilities' className='text-muted-foreground hover:text-foreground px-3 text-sm transition-colors'>
              See what it covers
            </Link>
          </div>
        </div>

        <div className='relative min-w-0'>
          <div className='bg-muted/40 absolute -inset-8 -z-10 rounded-[2rem] blur-3xl' />
          <div className='border-border/70 bg-card overflow-hidden rounded-2xl border shadow-2xl shadow-black/10'>
            <div className='border-border/70 flex items-center justify-between border-b px-5 py-4'>
              <div className='flex items-center gap-2'>
                <span className='bg-foreground text-background flex size-7 items-center justify-center rounded-lg'><Icons.logo className='size-3.5' /></span>
                <span className='text-sm font-semibold'>Overview</span>
              </div>
              <span className='text-muted-foreground text-xs'>Gonepost / Local</span>
            </div>
            <div className='grid gap-5 p-5 sm:p-7'>
              <div className='flex items-end justify-between'>
                <div><p className='text-muted-foreground text-xs uppercase'>Workspace pulse</p><p className='mt-2 text-3xl font-semibold'>Quietly in control</p></div>
                <span className='text-emerald-600 dark:text-emerald-400 text-xs font-medium'>All systems ready</span>
              </div>
              <div className='grid gap-3 sm:grid-cols-3'>
                {[['Users', '248', 'teams'], ['Roles', '12', 'lock'], ['Audit events', '1,482', 'clock']].map(([label, value, icon]) => {
                  const Icon = Icons[icon as keyof typeof Icons];
                  return <div key={label} className='bg-muted/50 rounded-xl p-4'><Icon className='text-muted-foreground size-4' /><p className='text-muted-foreground mt-5 text-xs'>{label}</p><p className='mt-1 text-xl font-semibold'>{value}</p></div>;
                })}
              </div>
              <div className='border-border/70 rounded-xl border p-4'>
                <div className='mb-5 flex items-center justify-between'><p className='text-sm font-medium'>Recent activity</p><span className='text-muted-foreground text-xs'>Live from API</span></div>
                <div className='space-y-4'>
                  {[['Role updated', 'roles:write', '2m'], ['User invited', 'users:write', '18m'], ['Access reviewed', 'audit:read', '41m']].map(([title, permission, time]) => <div key={title} className='flex items-center gap-3'><span className='bg-muted flex size-7 items-center justify-center rounded-lg'><Icons.check className='size-3.5' /></span><div className='min-w-0 flex-1'><p className='truncate text-sm'>{title}</p><p className='text-muted-foreground text-xs'>{permission}</p></div><span className='text-muted-foreground text-xs'>{time}</span></div>)}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id='capabilities' className='border-border/70 border-t'>
        <div className='mx-auto grid w-full max-w-7xl gap-10 px-6 py-20 lg:grid-cols-[0.7fr_1.3fr] lg:px-10'>
          <ScrollReveal>
            <p className='text-muted-foreground text-xs font-semibold tracking-[0.2em] uppercase'>One workspace, less noise</p>
            <h2 className='mt-4 max-w-sm text-3xl font-semibold tracking-tight sm:text-4xl'>Good operations are mostly good visibility.</h2>
          </ScrollReveal>
          <div className='grid gap-4 sm:grid-cols-3'>
            {capabilities.map((capability, i) => {
              const Icon = Icons[capability.icon];
              return (
                <ScrollReveal key={capability.title} delay={i * 120}>
                  <article className='border-border/70 h-full rounded-xl border p-5'>
                    <Icon className='text-muted-foreground size-5' />
                    <h3 className='mt-8 font-medium'>{capability.title}</h3>
                    <p className='text-muted-foreground mt-2 text-sm leading-6'>{capability.description}</p>
                  </article>
                </ScrollReveal>
              );
            })}
          </div>
        </div>
      </section>

      <footer className='border-border/70 border-t'>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 px-6 py-8 text-sm sm:flex-row sm:items-center sm:justify-between lg:px-10'>
          <span className='text-muted-foreground'>Gonepost / Operations workspace</span>
          <Link href='/auth/sign-in' className='text-foreground font-medium hover:underline'>Enter the workspace <span aria-hidden='true'>-&gt;</span></Link>
        </div>
      </footer>
    </main>
  );
}
