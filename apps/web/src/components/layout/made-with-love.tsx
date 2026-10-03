import { cn } from '@/lib/utils';

const GITHUB = 'https://github.com/alifma';

/** Small credit line: "Made with ♥ by alifma", linking to the GitHub profile. */
export function MadeWithLove({ className }: { className?: string }) {
  return (
    <p className={cn('text-muted-foreground text-xs', className)}>
      Made with <span className='text-red-500' role='img' aria-label='love'>&hearts;</span> by{' '}
      <a
        href={GITHUB}
        target='_blank'
        rel='noopener noreferrer'
        className='text-foreground font-medium underline-offset-4 hover:underline'
      >
        alifma
      </a>
    </p>
  );
}
