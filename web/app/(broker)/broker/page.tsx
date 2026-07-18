import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";

export default function BrokerDashboard() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-display font-bold text-slate-900 dark:text-slate-50">Broker Dashboard</h1>
          <p className="mt-1 text-slate-500 dark:text-slate-400">Manage your clients, quotes, and commissions.</p>
        </div>
        <Button variant="primary">New Quote</Button>
      </div>

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard title="Active Clients" value="142" trend="+12 this month" />
        <MetricCard title="Policies Bound" value="89" trend="+5 this month" />
        <MetricCard title="Pending Quotes" value="14" trend="Action required" />
        <MetricCard title="YTD Commission" value="ETB 124K" trend="On track" />
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Recent Quotes</CardTitle>
            <CardDescription>Latest generated quotations for your clients</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-left">
                <thead className="text-xs text-slate-500 uppercase bg-slate-50 dark:bg-slate-800 dark:text-slate-400">
                  <tr>
                    <th className="px-4 py-3 rounded-l-lg">Client</th>
                    <th className="px-4 py-3">Product</th>
                    <th className="px-4 py-3">Premium</th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3 rounded-r-lg">Action</th>
                  </tr>
                </thead>
                <tbody>
                  <TableRow client="Abebe Bekele" product="Commercial Motor" premium="ETB 15,000" status="Draft" />
                  <TableRow client="Nile Logistics" product="Goods-in-Transit" premium="ETB 45,000" status="Sent" />
                  <TableRow client="Selamawit T." product="Endowment Life" premium="ETB 8,500" status="Bound" />
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
        
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>Upcoming Renewals</CardTitle>
            <CardDescription>Policies expiring in 30 days</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <RenewalItem client="Kaleb Motors" date="12 Aug" />
              <RenewalItem client="Zemen Tech" date="15 Aug" />
              <RenewalItem client="Mekdes F." date="18 Aug" />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function MetricCard({ title, value, trend }: { title: string; value: string; trend: string }) {
  return (
    <Card>
      <CardHeader className="p-4 pb-2">
        <CardTitle className="text-sm font-medium text-slate-500 dark:text-slate-400">{title}</CardTitle>
      </CardHeader>
      <CardContent className="p-4 pt-0">
        <div className="text-2xl font-bold text-slate-900 dark:text-slate-50">{value}</div>
        <p className="text-xs text-brand-blue-600 dark:text-brand-gold mt-1 font-medium">{trend}</p>
      </CardContent>
    </Card>
  );
}

function TableRow({ client, product, premium, status }: { client: string; product: string; premium: string; status: string }) {
  return (
    <tr className="border-b border-slate-100 dark:border-slate-800 last:border-0">
      <td className="px-4 py-3 font-medium text-slate-900 dark:text-slate-100">{client}</td>
      <td className="px-4 py-3 text-slate-600 dark:text-slate-300">{product}</td>
      <td className="px-4 py-3 text-slate-600 dark:text-slate-300">{premium}</td>
      <td className="px-4 py-3 text-brand-blue-600 dark:text-brand-gold font-medium">{status}</td>
      <td className="px-4 py-3"><Button variant="ghost" size="sm">View</Button></td>
    </tr>
  );
}

function RenewalItem({ client, date }: { client: string; date: string }) {
  return (
    <div className="flex items-center justify-between border-b border-slate-100 pb-3 last:border-0 last:pb-0 dark:border-slate-800">
      <p className="font-medium text-slate-900 dark:text-slate-100">{client}</p>
      <span className="text-xs font-semibold px-2 py-1 bg-red-100 text-red-700 rounded-md dark:bg-red-900/30 dark:text-red-400">{date}</span>
    </div>
  );
}
