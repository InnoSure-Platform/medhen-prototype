import type { Meta, StoryObj } from "@storybook/react";
import { MoreHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const meta: Meta = { title: "Overlays/Gallery", parameters: { layout: "padded" } };
export default meta;
type Story = StoryObj;

export const DialogStory: Story = {
  name: "Dialog",
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button>Settle claim</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Settle claim (fast-track)</DialogTitle>
          <DialogDescription>Approve a settlement within your authority limit.</DialogDescription>
        </DialogHeader>
        <p className="text-sm text-fg-muted">Amount: <span className="font-mono font-semibold text-fg">42,000.00 ETB</span></p>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="secondary">Cancel</Button>
          </DialogClose>
          <Button>Confirm settlement</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  ),
};

export const Sheet: Story = {
  render: () => (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="secondary">Open details panel</Button>
      </DialogTrigger>
      <DialogContent variant="sheet">
        <DialogHeader>
          <DialogTitle>Policy EIC/MOT/2026/000001</DialogTitle>
          <DialogDescription>Read-only detail panel</DialogDescription>
        </DialogHeader>
        <div className="text-sm text-fg-muted">Slide-over sheet content…</div>
      </DialogContent>
    </Dialog>
  ),
};

export const Menu: Story = {
  render: () => (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="secondary" size="icon" aria-label="Actions">
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        <DropdownMenuLabel>Policy actions</DropdownMenuLabel>
        <DropdownMenuItem>View documents</DropdownMenuItem>
        <DropdownMenuItem>Download certificate</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem className="text-danger">Cancel policy</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  ),
};

export const TooltipStory: Story = {
  name: "Tooltip",
  render: () => (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="ghost">Combined ratio</Button>
        </TooltipTrigger>
        <TooltipContent>Loss ratio + assumed 30% expense ratio</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  ),
};

export const PopoverStory: Story = {
  name: "Popover",
  render: () => (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="secondary">Filters</Button>
      </PopoverTrigger>
      <PopoverContent>
        <p className="text-sm font-medium text-fg">Filter by status</p>
        <p className="mt-1 text-sm text-fg-muted">Popover body content…</p>
      </PopoverContent>
    </Popover>
  ),
};

export const TabsStory: Story = {
  name: "Tabs",
  render: () => (
    <Tabs defaultValue="overview" className="max-w-lg">
      <TabsList>
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="coverages">Coverages</TabsTrigger>
        <TabsTrigger value="documents">Documents</TabsTrigger>
      </TabsList>
      <TabsContent value="overview" className="text-sm text-fg-muted">Policy overview…</TabsContent>
      <TabsContent value="coverages" className="text-sm text-fg-muted">OD + TPL…</TabsContent>
      <TabsContent value="documents" className="text-sm text-fg-muted">Certificate of Insurance…</TabsContent>
    </Tabs>
  ),
};
