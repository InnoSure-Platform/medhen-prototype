import { Spinner } from "@/components/ui/spinner";

export default function Loading() {
  return (
    <div className="grid min-h-[60dvh] place-items-center">
      <Spinner className="size-7" />
    </div>
  );
}
