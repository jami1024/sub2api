<template>
  <div class="relative isolate overflow-hidden rounded-[2.25rem] border border-slate-200/70 bg-white/80 p-5 shadow-[0_24px_80px_rgba(15,23,42,0.10)] dark:border-white/10 dark:bg-[#0b1220]/88 dark:shadow-[0_30px_100px_rgba(2,6,23,0.65)] sm:p-7">
    <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(96,165,250,0.14),transparent_32%),radial-gradient(circle_at_top_right,rgba(251,146,60,0.10),transparent_24%)] dark:bg-[radial-gradient(circle_at_top_left,rgba(96,165,250,0.24),transparent_30%),radial-gradient(circle_at_top_right,rgba(251,146,60,0.16),transparent_24%)]"></div>
    <div class="absolute inset-0 opacity-20 [background-image:radial-gradient(rgba(15,23,42,0.12)_1px,transparent_1px)] [background-size:24px_24px] dark:opacity-25 dark:[background-image:radial-gradient(rgba(255,255,255,0.16)_1px,transparent_1px)]"></div>

    <div class="relative space-y-4 text-slate-900 dark:text-white">
      <div class="inline-flex rounded-full border border-slate-300/80 bg-white/75 px-4 py-2 text-[11px] font-semibold uppercase tracking-[0.28em] text-slate-500 dark:border-white/10 dark:bg-white/5 dark:text-slate-300">
        Stable Access
      </div>

      <div class="rounded-[1.8rem] border border-slate-200 bg-white/95 p-4 shadow-[0_18px_50px_rgba(15,23,42,0.08)] dark:border-white/10 dark:bg-white/95 dark:text-slate-900">
        <div class="flex items-center justify-between gap-3">
          <div class="h-3 w-24 rounded-full bg-slate-200"></div>
          <div class="rounded-xl bg-slate-950 px-4 py-2 text-xs font-semibold text-white">Always On</div>
        </div>
        <div class="mt-4 rounded-2xl border border-slate-200 bg-white px-4 py-4 text-sm text-slate-500 shadow-inner">
          AIGO.RUN / stable route
        </div>
      </div>

      <div class="grid gap-3 sm:grid-cols-3">
        <div class="rounded-2xl border border-slate-200 bg-white/80 px-4 py-4 text-sm font-medium text-slate-700 dark:border-white/10 dark:bg-white/5 dark:text-slate-200">稳定使用</div>
        <div class="rounded-2xl border border-slate-200 bg-white/80 px-4 py-4 text-sm font-medium text-slate-700 dark:border-white/10 dark:bg-white/5 dark:text-slate-200">省心接入</div>
        <div class="rounded-2xl border border-slate-200 bg-white/80 px-4 py-4 text-sm font-medium text-slate-700 dark:border-white/10 dark:bg-white/5 dark:text-slate-200">长期可用</div>
      </div>
    </div>
  </div>
</template>
