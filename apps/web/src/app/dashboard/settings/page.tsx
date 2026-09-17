import PageContainer from '@/components/layout/page-container';
import { ThemeModeToggle } from '@/components/themes/theme-mode-toggle';
import { ThemeSelector } from '@/components/themes/theme-selector';

export const metadata = { title: 'Dashboard: Settings' };

export default function SettingsPage() {
  return (
    <PageContainer
      pageTitle='Settings'
      pageDescription='Tune the dashboard appearance for your local workspace.'
    >
      <div className='grid max-w-3xl gap-6'>
        <section className='border-border/70 bg-card rounded-xl border p-6'>
          <div>
            <h2 className='font-semibold'>Appearance</h2>
            <p className='text-muted-foreground mt-1 text-sm'>Choose the visual theme and color mode.</p>
          </div>
          <div className='border-border/70 mt-6 grid gap-5 border-t pt-5 sm:grid-cols-[1fr_auto] sm:items-center'>
            <div>
              <h3 className='text-sm font-medium'>Theme family</h3>
              <p className='text-muted-foreground text-sm'>Select the dashboard color system.</p>
            </div>
            <ThemeSelector />
          </div>
          <div className='border-border/70 mt-5 grid gap-5 border-t pt-5 sm:grid-cols-[1fr_auto] sm:items-center'>
            <div>
              <h3 className='text-sm font-medium'>Color mode</h3>
              <p className='text-muted-foreground text-sm'>Switch between light and dark mode.</p>
            </div>
            <ThemeModeToggle />
          </div>
        </section>
      </div>
    </PageContainer>
  );
}