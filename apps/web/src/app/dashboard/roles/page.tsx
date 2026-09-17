import PageContainer from '@/components/layout/page-container';
import ApiRolesPanel from '@/features/roles/components/api-roles-panel';

export const metadata = { title: 'Dashboard: Roles' };

export default function RolesPage() {
  return (
    <PageContainer pageTitle='Roles' pageDescription='Review roles and permissions from the Gonepost API.'>
      <ApiRolesPanel />
    </PageContainer>
  );
}