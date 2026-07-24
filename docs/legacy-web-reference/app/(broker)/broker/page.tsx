import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";

export default function BrokerDashboard() {
  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-8 bg-white border border-slate-200 rounded-2xl shadow-sm relative overflow-hidden">
        <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-bl from-brand-blue-50 to-transparent rounded-full blur-3xl -z-10" />
        <div>
          <div className="inline-flex items-center gap-2 px-3 py-1 mb-3 rounded-full bg-blue-50 text-brand-blue-600 text-xs font-bold tracking-wider uppercase border border-blue-100">
            <span className="w-1.5 h-1.5 rounded-full bg-brand-blue-600 animate-pulse" />
            Live System
          </div>
          <h1 className="text-3xl font-display font-bold text-slate-900 tracking-tight">Broker Dashboard</h1>
          <p className="mt-1.5 text-slate-500 max-w-lg">Manage your client portfolio, generate new quotes, and track your commission performance across all lines of business.</p>
        </div>
        <button className="flex items-center gap-2 bg-brand-blue-600 hover:bg-brand-blue-500 text-white px-6 py-3 rounded-xl font-semibold transition-all shadow-lg shadow-brand-blue-600/20 hover:shadow-xl hover:-translate-y-0.5 active:translate-y-0">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" /></svg>
          New Quote
        </button>
      </div>

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard title="Active Clients" value="142" trend="+12 this month" trendType="positive" icon="users" />
        <MetricCard title="Policies Bound" value="89" trend="+5 this month" trendType="positive" icon="shield" />
        <MetricCard title="Pending Quotes" value="14" trend="Action required" trendType="warning" icon="file" />
        <MetricCard title="YTD Commission" value="ETB 124K" trend="On track" trendType="neutral" icon="chart" />
      </div>

      <div className="grid gap-8 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold text-slate-900 font-display">Recent Quotes</h2>
            <button className="text-sm font-medium text-brand-blue-600 hover:text-brand-blue-800 transition-colors">View All →</button>
          </div>
          <div className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-sm text-left">
                <thead className="text-xs text-slate-500 uppercase bg-slate-50/80 border-b border-slate-200">
                  <tr>
                    <th className="px-6 py-4 font-semibold tracking-wider">Client</th>
                    <th className="px-6 py-4 font-semibold tracking-wider">Product</th>
                    <th className="px-6 py-4 font-semibold tracking-wider">Premium</th>
                    <th className="px-6 py-4 font-semibold tracking-wider">Status</th>
                    <th className="px-6 py-4 font-semibold tracking-wider text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  <TableRow client="Abebe Bekele" product="Commercial Motor" premium="ETB 15,000" status="Draft" />
                  <TableRow client="Nile Logistics" product="Goods-in-Transit" premium="ETB 45,000" status="Sent" />
                  <TableRow client="Selamawit T." product="Endowment Life" premium="ETB 8,500" status="Bound" />
                  <TableRow client="Addis Manufacturing" product="Fire & Lightning" premium="ETB 120,000" status="Referred" />
                </tbody>
              </table>
            </div>
          </div>
        </div>
        
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold text-slate-900 font-display">Upcoming Renewals</h2>
          </div>
          <div className="bg-white border border-slate-200 rounded-2xl shadow-sm p-1">
            <div className="divide-y divide-slate-100">
              <RenewalItem client="Kaleb Motors" date="12 Aug" daysLeft={4} />
              <RenewalItem client="Zemen Tech" date="15 Aug" daysLeft={7} />
              <RenewalItem client="Mekdes F." date="18 Aug" daysLeft={10} />
              <RenewalItem client="Pioneer Trading" date="22 Aug" daysLeft={14} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function MetricCard({ title, value, trend, trendType, icon }: { title: string; value: string; trend: string; trendType: 'positive'|'warning'|'neutral'; icon: string }) {
  const trendColors = {
    positive: "text-emerald-600 bg-emerald-50 border-emerald-100",
    warning: "text-amber-600 bg-amber-50 border-amber-100",
    neutral: "text-brand-blue-600 bg-blue-50 border-blue-100"
  };

  const IconMap = {
    users: <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />,
    shield: <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />,
    file: <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />,
    chart: <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
  };

  return (
    <div className="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-300 relative overflow-hidden group">
      <div className="absolute -right-4 -top-4 w-24 h-24 bg-slate-50 rounded-full group-hover:scale-150 transition-transform duration-700 ease-out z-0" />
      <div className="relative z-10">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-semibold text-slate-500 uppercase tracking-wider">{title}</h3>
          <div className="p-2 bg-slate-50 rounded-lg text-slate-400 group-hover:text-brand-blue-600 transition-colors border border-slate-100">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              {IconMap[icon as keyof typeof IconMap]}
            </svg>
          </div>
        </div>
        <div className="text-3xl font-bold text-slate-900 font-display mb-3">{value}</div>
        <div className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold border ${trendColors[trendType]}`}>
          {trend}
        </div>
      </div>
    </div>
  );
}

function TableRow({ client, product, premium, status }: { client: string; product: string; premium: string; status: string }) {
  const statusColors: Record<string, string> = {
    "Draft": "bg-slate-100 text-slate-600 border-slate-200",
    "Sent": "bg-blue-50 text-blue-700 border-blue-200",
    "Bound": "bg-emerald-50 text-emerald-700 border-emerald-200",
    "Referred": "bg-amber-50 text-amber-700 border-amber-200"
  };

  return (
    <tr className="hover:bg-slate-50/80 transition-colors group">
      <td className="px-6 py-4">
        <div className="font-semibold text-slate-900">{client}</div>
        <div className="text-xs text-slate-500 mt-0.5">ID: {Math.floor(Math.random()*10000).toString().padStart(5, '0')}</div>
      </td>
      <td className="px-6 py-4">
        <div className="text-slate-700 font-medium">{product}</div>
      </td>
      <td className="px-6 py-4">
        <div className="text-slate-900 font-medium font-mono">{premium}</div>
      </td>
      <td className="px-6 py-4">
        <span className={`px-2.5 py-1 text-xs font-semibold rounded-full border ${statusColors[status] || statusColors["Draft"]}`}>
          {status}
        </span>
      </td>
      <td className="px-6 py-4 text-right">
        <button className="opacity-0 group-hover:opacity-100 transition-opacity text-sm font-semibold text-brand-blue-600 hover:text-brand-blue-800 bg-blue-50 hover:bg-blue-100 px-3 py-1.5 rounded-lg">
          View Details
        </button>
      </td>
    </tr>
  );
}

function RenewalItem({ client, date, daysLeft }: { client: string; date: string; daysLeft: number }) {
  return (
    <div className="flex items-center justify-between p-4 hover:bg-slate-50 rounded-xl transition-colors cursor-pointer group">
      <div className="flex items-center gap-4">
        <div className={`w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm ${daysLeft <= 5 ? 'bg-red-50 text-red-600 border border-red-100' : 'bg-amber-50 text-amber-600 border border-amber-100'}`}>
          {daysLeft}d
        </div>
        <div>
          <p className="font-semibold text-slate-900 group-hover:text-brand-blue-600 transition-colors">{client}</p>
          <p className="text-xs text-slate-500 font-medium">Exp: {date}</p>
        </div>
      </div>
      <svg className="w-5 h-5 text-slate-300 group-hover:text-brand-blue-600 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" /></svg>
    </div>
  );
}
