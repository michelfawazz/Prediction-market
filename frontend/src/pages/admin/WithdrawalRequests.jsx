import React, { useState, useEffect, useCallback } from 'react';
import { API_URL } from '../../config';

const WithdrawalRequests = () => {
    const [withdrawals, setWithdrawals] = useState([]);
    const [stats, setStats] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [filter, setFilter] = useState('PENDING');
    const [selectedWithdrawal, setSelectedWithdrawal] = useState(null);
    const [actionLoading, setActionLoading] = useState(false);
    const [rejectReason, setRejectReason] = useState('');

    const fetchWithdrawals = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const token = localStorage.getItem('token');
            const url = filter
                ? `${API_URL}/v0/admin/withdrawals?status=${filter}`
                : `${API_URL}/v0/admin/withdrawals`;

            const response = await fetch(url, {
                headers: { 'Authorization': `Bearer ${token}` },
            });

            if (!response.ok) {
                throw new Error('Failed to fetch withdrawals');
            }

            const data = await response.json();
            setWithdrawals(data.withdrawals || []);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }, [filter]);

    const fetchStats = useCallback(async () => {
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${API_URL}/v0/admin/withdrawals/stats`, {
                headers: { 'Authorization': `Bearer ${token}` },
            });

            if (response.ok) {
                const data = await response.json();
                setStats(data);
            }
        } catch (err) {
            console.error('Failed to fetch stats:', err);
        }
    }, []);

    useEffect(() => {
        fetchWithdrawals();
        fetchStats();
    }, [fetchWithdrawals, fetchStats]);

    const handleApprove = async (withdrawalId) => {
        setActionLoading(true);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${API_URL}/v0/admin/withdrawals/${withdrawalId}/approve`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({}),
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Failed to approve withdrawal');
            }

            setSelectedWithdrawal(null);
            fetchWithdrawals();
            fetchStats();
        } catch (err) {
            alert(`Error: ${err.message}`);
        } finally {
            setActionLoading(false);
        }
    };

    const handleReject = async (withdrawalId) => {
        if (!rejectReason.trim()) {
            alert('Please provide a reason for rejection');
            return;
        }

        setActionLoading(true);
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${API_URL}/v0/admin/withdrawals/${withdrawalId}/reject`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ reason: rejectReason }),
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Failed to reject withdrawal');
            }

            setSelectedWithdrawal(null);
            setRejectReason('');
            fetchWithdrawals();
            fetchStats();
        } catch (err) {
            alert(`Error: ${err.message}`);
        } finally {
            setActionLoading(false);
        }
    };

    const getStatusColor = (status) => {
        switch (status) {
            case 'COMPLETED':
                return 'bg-green-600';
            case 'PENDING':
                return 'bg-yellow-600';
            case 'APPROVED':
                return 'bg-blue-600';
            case 'FAILED':
            case 'REJECTED':
                return 'bg-red-600';
            default:
                return 'bg-gray-600';
        }
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleString('en-US', {
            month: 'short',
            day: 'numeric',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const truncateAddress = (address) => {
        if (!address) return '';
        return `${address.slice(0, 10)}...${address.slice(-8)}`;
    };

    return (
        <div className="p-4 space-y-6">
            {/* Stats Overview */}
            {stats && (
                <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                    <div className="bg-yellow-900/30 border border-yellow-600/50 rounded-lg p-4">
                        <p className="text-yellow-400 text-2xl font-bold">{stats.pending?.count || 0}</p>
                        <p className="text-gray-400 text-sm">Pending</p>
                        <p className="text-yellow-400 text-xs">{stats.pending?.amount || 0} credits</p>
                    </div>
                    <div className="bg-blue-900/30 border border-blue-600/50 rounded-lg p-4">
                        <p className="text-blue-400 text-2xl font-bold">{stats.approved?.count || 0}</p>
                        <p className="text-gray-400 text-sm">Processing</p>
                    </div>
                    <div className="bg-green-900/30 border border-green-600/50 rounded-lg p-4">
                        <p className="text-green-400 text-2xl font-bold">{stats.completed?.count || 0}</p>
                        <p className="text-gray-400 text-sm">Completed</p>
                        <p className="text-green-400 text-xs">{stats.completed?.amount || 0} credits</p>
                    </div>
                    <div className="bg-red-900/30 border border-red-600/50 rounded-lg p-4">
                        <p className="text-red-400 text-2xl font-bold">{stats.rejected?.count || 0}</p>
                        <p className="text-gray-400 text-sm">Rejected</p>
                    </div>
                    <div className="bg-gray-900/30 border border-gray-600/50 rounded-lg p-4">
                        <p className="text-gray-400 text-2xl font-bold">{stats.failed?.count || 0}</p>
                        <p className="text-gray-400 text-sm">Failed</p>
                    </div>
                </div>
            )}

            {/* Filters */}
            <div className="flex flex-wrap gap-2">
                {['PENDING', 'APPROVED', 'COMPLETED', 'REJECTED', 'FAILED', ''].map((status) => (
                    <button
                        key={status || 'all'}
                        onClick={() => setFilter(status)}
                        className={`px-4 py-2 rounded-lg text-sm transition-colors
                            ${filter === status
                                ? 'bg-blue-600 text-white'
                                : 'bg-gray-700 text-gray-300 hover:bg-gray-600'}`}
                    >
                        {status || 'All'}
                    </button>
                ))}
                <button
                    onClick={() => { fetchWithdrawals(); fetchStats(); }}
                    className="px-4 py-2 bg-gray-700 text-gray-300 rounded-lg hover:bg-gray-600 ml-auto"
                >
                    Refresh
                </button>
            </div>

            {/* Withdrawals List */}
            {loading ? (
                <div className="flex justify-center py-8">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-white"></div>
                </div>
            ) : error ? (
                <div className="text-center py-8">
                    <p className="text-red-400">{error}</p>
                </div>
            ) : withdrawals.length === 0 ? (
                <div className="text-center py-8">
                    <p className="text-gray-400">No withdrawal requests found</p>
                </div>
            ) : (
                <div className="overflow-x-auto">
                    <table className="w-full">
                        <thead>
                            <tr className="text-left text-gray-400 text-sm border-b border-gray-700">
                                <th className="pb-3 pr-4">ID</th>
                                <th className="pb-3 pr-4">User</th>
                                <th className="pb-3 pr-4">Amount</th>
                                <th className="pb-3 pr-4">Token</th>
                                <th className="pb-3 pr-4">Chain</th>
                                <th className="pb-3 pr-4">To Address</th>
                                <th className="pb-3 pr-4">Status</th>
                                <th className="pb-3 pr-4">Date</th>
                                <th className="pb-3">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {withdrawals.map((w) => (
                                <tr key={w.id} className="border-b border-gray-800 hover:bg-gray-800/50">
                                    <td className="py-3 pr-4 text-white">#{w.id}</td>
                                    <td className="py-3 pr-4 text-white">{w.username}</td>
                                    <td className="py-3 pr-4 text-white font-mono">{w.amount}</td>
                                    <td className="py-3 pr-4 text-gray-300">{w.tokenSymbol}</td>
                                    <td className="py-3 pr-4 text-gray-300 capitalize">{w.chainName}</td>
                                    <td className="py-3 pr-4 text-gray-300 font-mono text-xs">
                                        {truncateAddress(w.toAddress)}
                                    </td>
                                    <td className="py-3 pr-4">
                                        <span className={`px-2 py-1 rounded text-xs text-white ${getStatusColor(w.status)}`}>
                                            {w.status}
                                        </span>
                                    </td>
                                    <td className="py-3 pr-4 text-gray-400 text-sm">
                                        {formatDate(w.createdAt)}
                                    </td>
                                    <td className="py-3">
                                        {w.status === 'PENDING' && (
                                            <button
                                                onClick={() => setSelectedWithdrawal(w)}
                                                className="text-blue-400 hover:text-blue-300 text-sm"
                                            >
                                                Review
                                            </button>
                                        )}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Review Modal */}
            {selectedWithdrawal && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center z-50">
                    <div className="bg-gray-800 rounded-lg w-full max-w-md mx-4 p-6">
                        <h3 className="text-xl font-bold text-white mb-4">
                            Review Withdrawal #{selectedWithdrawal.id}
                        </h3>

                        <div className="space-y-3 mb-6">
                            <div className="flex justify-between">
                                <span className="text-gray-400">User:</span>
                                <span className="text-white">{selectedWithdrawal.username}</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-400">Amount:</span>
                                <span className="text-white font-mono">
                                    {selectedWithdrawal.amount} {selectedWithdrawal.tokenSymbol}
                                </span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-400">Chain:</span>
                                <span className="text-white capitalize">{selectedWithdrawal.chainName}</span>
                            </div>
                            <div>
                                <span className="text-gray-400">To Address:</span>
                                <p className="text-white font-mono text-sm break-all mt-1">
                                    {selectedWithdrawal.toAddress}
                                </p>
                            </div>
                        </div>

                        {/* Reject Reason Input */}
                        <div className="mb-4">
                            <label className="block text-gray-400 text-sm mb-2">
                                Rejection Reason (required for reject)
                            </label>
                            <textarea
                                value={rejectReason}
                                onChange={(e) => setRejectReason(e.target.value)}
                                placeholder="Enter reason if rejecting..."
                                className="w-full bg-gray-700 text-white rounded-lg px-3 py-2 border border-gray-600 focus:border-blue-500 focus:outline-none"
                                rows={2}
                            />
                        </div>

                        <div className="flex space-x-3">
                            <button
                                onClick={() => handleApprove(selectedWithdrawal.id)}
                                disabled={actionLoading}
                                className="flex-1 bg-green-600 hover:bg-green-700 disabled:bg-gray-600 text-white py-2 rounded-lg"
                            >
                                {actionLoading ? 'Processing...' : 'Approve & Send'}
                            </button>
                            <button
                                onClick={() => handleReject(selectedWithdrawal.id)}
                                disabled={actionLoading}
                                className="flex-1 bg-red-600 hover:bg-red-700 disabled:bg-gray-600 text-white py-2 rounded-lg"
                            >
                                {actionLoading ? 'Processing...' : 'Reject & Refund'}
                            </button>
                        </div>

                        <button
                            onClick={() => {
                                setSelectedWithdrawal(null);
                                setRejectReason('');
                            }}
                            className="w-full mt-3 text-gray-400 hover:text-white text-sm"
                        >
                            Cancel
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default WithdrawalRequests;
