import PageContainer from '@/components/layout/page-container';
import ApiAuditLogPanel from '@/features/audit-logs/components/api-audit-log-panel';

export const metadata = { title: 'Dashboard: Audit Logs' };

export default function AuditLogsPage() {
  return (
    <PageContainer pageTitle='Audit Logs' pageDescription='Inspect security and mutation events from the Gonepost API.'>
      <ApiAuditLogPanel />
    </PageContainer>
  );
}