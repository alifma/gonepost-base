import { delay } from '@/lib/delay';
import { AreaGraph } from '@/features/overview/components/area-graph';

export default async function AreaStats() {
  await delay(2000);
  return <AreaGraph />;
}
