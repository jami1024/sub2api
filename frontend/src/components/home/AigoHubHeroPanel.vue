<template>
  <div class="relative isolate overflow-hidden rounded-[2rem] border border-white/10 bg-[#0b1220]/90 p-6 shadow-[0_30px_100px_rgba(2,6,23,0.65)] sm:p-8">
    <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(96,165,250,0.24),transparent_30%),radial-gradient(circle_at_top_right,rgba(251,146,60,0.16),transparent_24%)]"></div>
    <div class="absolute inset-0 opacity-25 [background-image:radial-gradient(rgba(255,255,255,0.16)_1px,transparent_1px)] [background-size:22px_22px]"></div>

    <div class="relative space-y-5 text-white">
      <div class="inline-flex rounded-full border border-white/10 bg-white/5 px-4 py-2 text-[11px] font-semibold uppercase tracking-[0.28em] text-slate-300">
        Stable Access
      </div>

      <div class="space-y-3">
        <div class="text-3xl font-black leading-tight text-white sm:text-4xl">
          接入更简单，
          <span class="bg-gradient-to-r from-sky-300 via-blue-300 to-indigo-300 bg-clip-text text-transparent">
            稳定才是底线
          </span>
        </div>
        <p class="max-w-md text-sm leading-7 text-slate-300">
          AigoHub 的首屏展示只负责传达稳定感，不伪装成复杂后台，不堆不可信的假数据。
        </p>
      </div>

      <div class="grid gap-3 sm:grid-cols-[1.2fr_0.8fr]">
        <div class="rounded-[1.6rem] border border-white/10 bg-white/95 p-4 text-slate-900 shadow-[0_24px_60px_rgba(15,23,42,0.35)]">
          <div class="flex items-center justify-between gap-3">
            <div class="h-3 w-24 rounded-full bg-slate-200"></div>
            <div class="rounded-xl bg-gradient-to-r from-blue-500 to-indigo-500 px-4 py-2 text-xs font-semibold text-white">
              Always On
            </div>
          </div>
          <div class="mt-4 rounded-2xl border border-slate-200 bg-white px-4 py-4 text-sm text-slate-500 shadow-inner">
            AIGO.RUN / stable route
          </div>
          <div class="mt-4 grid grid-cols-2 gap-2">
            <div class="rounded-2xl bg-[#1f2a44] px-3 py-3 text-xs font-semibold text-slate-100">稳定使用</div>
            <div class="rounded-2xl bg-[#29498f] px-3 py-3 text-xs font-semibold text-slate-100">省心接入</div>
            <div class="rounded-2xl bg-[#253250] px-3 py-3 text-xs font-semibold text-slate-100">长期可用</div>
            <div class="rounded-2xl bg-[#31405f] px-3 py-3 text-xs font-semibold text-slate-100">联系即聊</div>
          </div>
        </div>

        <div class="grid gap-3">
          <div class="rounded-[1.4rem] border border-sky-400/20 bg-sky-400/10 p-4 text-sm text-sky-100 shadow-[0_18px_40px_rgba(56,189,248,0.14)]">
            Stable to use
          </div>
          <div class="rounded-[1.4rem] border border-white/10 bg-white/5 p-4 text-sm text-slate-200">
            Less friction
          </div>
          <div class="rounded-[1.4rem] border border-white/10 bg-white/5 p-4 text-sm text-slate-200">
            WeChat first
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
