import { apiClient } from '@/api/client'
import type { AffiliateWithdrawalRequest } from '@/types'

export async function getAffiliateWithdrawals(status = ''): Promise<AffiliateWithdrawalRequest[]> {
  const { data } = await apiClient.get<AffiliateWithdrawalRequest[]>('/admin/affiliate/withdrawals', {
    params: status ? { status } : undefined,
  })
  return data
}

export async function rejectAffiliateWithdrawal(id: number, admin_note = ''): Promise<AffiliateWithdrawalRequest> {
  const { data } = await apiClient.post<AffiliateWithdrawalRequest>(`/admin/affiliate/withdrawals/${id}/reject`, {
    admin_note,
  })
  return data
}

export async function markAffiliateWithdrawalPaid(id: number, admin_note = ''): Promise<AffiliateWithdrawalRequest> {
  const { data } = await apiClient.post<AffiliateWithdrawalRequest>(`/admin/affiliate/withdrawals/${id}/mark-paid`, {
    admin_note,
  })
  return data
}

export const adminAffiliateAPI = {
  getAffiliateWithdrawals,
  rejectAffiliateWithdrawal,
  markAffiliateWithdrawalPaid,
}

export default adminAffiliateAPI
