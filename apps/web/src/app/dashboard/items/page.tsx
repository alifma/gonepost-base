import PageContainer from '@/components/layout/page-container';
import ItemsPage from '@/features/items/components/items-page';

export const metadata = { title: 'Dashboard: Items' };

export default function Page() {
  return (
    <PageContainer pageTitle='Items' pageDescription='Reference CRUD feature. Copy it to start a new one.'>
      <ItemsPage />
    </PageContainer>
  );
}
