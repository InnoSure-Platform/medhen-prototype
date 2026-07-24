import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";

export default function CustomerDashboard() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-display font-bold text-slate-900 dark:text-slate-50">Welcome back, Sarah</h1>
        <p className="mt-1 text-slate-500 dark:text-slate-400">Here is an overview of your insurance portfolio.</p>
      </div>

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard title="Active Policies" value="3" />
        <MetricCard title="Pending Claims" value="1" />
        <MetricCard title="Next Premium Due" value="15 Aug, 2026" />
        <MetricCard title="Total Coverage" value="ETB 4.2M" />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>Your latest transactions and updates</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <ActivityItem title="Claim #CL-2026-89" status="In Review" date="Today" />
              <ActivityItem title="Motor Policy Renewal" status="Paid" date="2 days ago" />
              <ActivityItem title="Life Insurance Premium" status="Paid" date="Last week" />
            </div>
          </CardContent>
        </Card>
        
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Frequently used services</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4 sm:grid-cols-2">
            <Button variant="outline" className="h-20 flex-col gap-2">
              <span>File a Claim</span>
            </Button>
            <Button variant="outline" className="h-20 flex-col gap-2">
              <span>Pay Premium</span>
            </Button>
            <Button variant="outline" className="h-20 flex-col gap-2">
              <span>Get a Quote</span>
            </Button>
            <Button variant="outline" className="h-20 flex-col gap-2">
              <span>Download ID Card</span>
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function MetricCard({ title, value }: { title: string; value: string }) {
  return (
    <Card>
      <CardHeader className="p-4 pb-2">
        <CardTitle className="text-sm font-medium text-slate-500 dark:text-slate-400">{title}</CardTitle>
      </CardHeader>
      <CardContent className="p-4 pt-0">
        <div className="text-2xl font-bold text-slate-900 dark:text-slate-50">{value}</div>
      </CardContent>
    </Card>
  );
}

function ActivityItem({ title, status, date }: { title: string; status: string; date: string }) {
  return (
    <div className="flex items-center justify-between border-b border-slate-100 pb-4 last:border-0 last:pb-0 dark:border-slate-800">
      <div>
        <p className="font-medium text-slate-900 dark:text-slate-100">{title}</p>
        <p className="text-sm text-slate-500 dark:text-slate-400">{date}</p>
      </div>
      <div className="text-sm font-medium text-brand-blue-600 dark:text-brand-gold">{status}</div>
    </div>
  );
}
