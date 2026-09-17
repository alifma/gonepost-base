import { delay } from '@/lib/delay';
import { BarGraph } from '@/features/overview/components/bar-graph';

export default async function BarStats() {
  await delay(1000);

  return <BarGraph />;
}
