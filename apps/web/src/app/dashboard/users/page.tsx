import PageContainer from '@/components/layout/page-container';
import ApiUsersPanel from '@/features/users/components/api-users-panel';

export const metadata = {
  title: 'Dashboard: Users'
};

export default function UsersPage() {
  return (
    <PageContainer
      pageTitle='Users'
      pageDescription='Manage users through the Gonepost API.'
    >
      <ApiUsersPanel />
    </PageContainer>
  );
}
