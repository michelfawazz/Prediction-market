import React, { useState, useEffect } from 'react';
import { API_URL } from '../../config';

const TransactionHistory = () => {
    const [transactions, setTransactions] = useState([]);
    const [withdrawals, setWithdrawals] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [activeView, setActiveView] = useState('all'); // 'all', 'deposits', 'withdrawals'

    useEffect(() => {
        fetchTransactions();
    }, []);

    const fetchTransactions = async () => {
        setLoading(true);
        setError(null);
        try {
            const token = localStorage.getItem('token');

            // Fetch both transactions and withdrawal requests
            const [txResponse, withdrawalResponse] = await Promise.all([
                fetch(`${API_URL}/v0/wallet/transactions`, {
                    headers: { 'Authorization': `Bearer ${token}` },
                }),
                fetch(`${API_URL}/v0/wallet/withdrawals`, {
                    headers: { 'Authorization': `Bearer ${token}` },
                }),
            ]);

            if (txResponse.ok) {
                const txData = await txResponse.json();
                setTransactions(txData.transactions || []);
            }

            if (withdrawalResponse.ok) {
                const withdrawalData = await withdrawalResponse.json();
                setWithdrawals(withdrawalData.withdrawals || []);
            }
        } catch (err) {
            setError('Failed to load transaction history');
        } finally {
            setLoading(false);
        }
    };

    const getStatusColor = (status) => {
        switch (status) {
            case 'COMPLETED':
                return 'text-green-400';
            case 'PENDING':
                return 'text-yellow-400';
            case 'APPROVED':
                return 'text-blue-400';
            case 'FAILED':
            case 'REJECTED':
                return 'text-red-400';
            default:
                return 'text-gray-400';
        }
    };

    const getTypeIcon = (type) => {
        if (type === 'DEPOSIT') {
            return (
                <span className="text-green-400">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8l-8 8-8-8" />
                    </svg>
                </span>
            );
        }
        return (
            <span className="text-red-400">
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 20V4m-8 8l8-8 8 8" />
                </svg>
            </span>
        );
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const truncateAddress = (address) => {
        if (!address) return '';
        return `${address.slice(0, 6)}...${address.slice(-4)}`;
    };

    // Combine and filter items based on view
    const getDisplayItems = () => {
        let items = [];

        if (activeView === 'all' || activeView === 'deposits') {
            const deposits = transactions
                .filter(tx => tx.type === 'DEPOSIT')
                .map(tx => ({ ...tx, source: 'transaction' }));
            items = [...items, ...deposits];
        }

        if (activeView === 'all' || activeView === 'withdrawals') {
            const withdrawalItems = withdrawals.map(w => ({
                id: `w-${w.id}`,
                type: 'WITHDRAWAL',
                status: w.status,
                chainName: w.chainName,
                tokenSymbol: w.tokenSymbol,
                amount: w.amount,
                toAddress: w.toAddress,
                createdAt: w.createdAt,
                source: 'withdrawal',
            }));
            items = [...items, ...withdrawalItems];
        }

        // Sort by date descending
        items.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

        return items;
    };

    const displayItems = getDisplayItems();

    if (loading) {
        return (
            <div className="flex justify-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-white"></div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="text-center py-8">
                <p className="text-red-400 mb-2">{error}</p>
                <button
                    onClick={fetchTransactions}
                    className="text-blue-400 hover:text-blue-300 text-sm"
                >
                    Try again
                </button>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            {/* Filter Tabs */}
            <div className="flex space-x-2">
                {['all', 'deposits', 'withdrawals'].map((view) => (
                    <button
                        key={view}
                        onClick={() => setActiveView(view)}
                        className={`px-3 py-1 text-sm rounded-lg transition-colors capitalize
                            ${activeView === view
                                ? 'bg-blue-600 text-white'
                                : 'bg-gray-700 text-gray-400 hover:text-white'}`}
                    >
                        {view}
                    </button>
                ))}
            </div>

            {/* Transaction List */}
            {displayItems.length === 0 ? (
                <div className="text-center py-8">
                    <p className="text-gray-400">No transactions yet</p>
                </div>
            ) : (
                <div className="space-y-2">
                    {displayItems.map((item) => (
                        <div
                            key={item.id}
                            className="bg-gray-700 rounded-lg p-3 flex items-center justify-between"
                        >
                            <div className="flex items-center space-x-3">
                                {getTypeIcon(item.type)}
                                <div>
                                    <div className="flex items-center space-x-2">
                                        <span className="text-white font-medium">
                                            {item.type === 'DEPOSIT' ? '+' : '-'}{item.amount}
                                        </span>
                                        <span className="text-gray-400 text-sm">
                                            {item.tokenSymbol}
                                        </span>
                                    </div>
                                    <div className="text-gray-500 text-xs">
                                        {item.chainName} · {formatDate(item.createdAt)}
                                    </div>
                                </div>
                            </div>
                            <div className="text-right">
                                <span className={`text-sm ${getStatusColor(item.status)}`}>
                                    {item.status}
                                </span>
                                {item.toAddress && (
                                    <div className="text-gray-500 text-xs">
                                        To: {truncateAddress(item.toAddress)}
                                    </div>
                                )}
                                {item.txHash && (
                                    <div className="text-gray-500 text-xs">
                                        {truncateAddress(item.txHash)}
                                    </div>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {/* Refresh Button */}
            <div className="text-center">
                <button
                    onClick={fetchTransactions}
                    className="text-blue-400 hover:text-blue-300 text-sm"
                >
                    Refresh
                </button>
            </div>
        </div>
    );
};

export default TransactionHistory;
