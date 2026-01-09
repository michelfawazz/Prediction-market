import React, { useState, useEffect } from 'react';
import { API_URL } from '../../config';
import { useAuth } from '../../helpers/AuthContent';
import ChainSelector from './ChainSelector';
import TokenSelector from './TokenSelector';

const WithdrawTab = ({ onClose }) => {
    const { username } = useAuth();
    const [userCredit, setUserCredit] = useState(0);
    const [selectedChain, setSelectedChain] = useState('ethereum');
    const [selectedToken, setSelectedToken] = useState('USDC');
    const [amount, setAmount] = useState('');
    const [toAddress, setToAddress] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [success, setSuccess] = useState(false);

    // Fetch user credit on mount
    useEffect(() => {
        const fetchCredit = async () => {
            try {
                const token = localStorage.getItem('token');
                const response = await fetch(`${API_URL}/v0/usercredit/${username}`, {
                    headers: {
                        'Authorization': `Bearer ${token}`,
                    },
                });
                if (response.ok) {
                    const data = await response.json();
                    setUserCredit(data.userCredit || 0);
                }
            } catch (err) {
                console.error('Failed to fetch credit:', err);
            }
        };
        if (username) {
            fetchCredit();
        }
    }, [username]);

    const handleWithdraw = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError(null);

        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${API_URL}/v0/wallet/withdraw`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    chainName: selectedChain,
                    tokenSymbol: selectedToken,
                    amount: parseInt(amount, 10),
                    toAddress: toAddress,
                }),
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Withdrawal failed');
            }

            setSuccess(true);
            setTimeout(() => {
                if (onClose) onClose();
            }, 3000);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const handleMaxAmount = () => {
        setAmount(userCredit.toString());
    };

    const isValidAddress = (addr) => {
        return /^0x[a-fA-F0-9]{40}$/.test(addr);
    };

    const canSubmit = amount && parseInt(amount, 10) >= 10 && toAddress && isValidAddress(toAddress) && !loading;

    if (success) {
        return (
            <div className="text-center py-8">
                <div className="text-green-400 text-6xl mb-4">
                    <svg className="w-16 h-16 mx-auto" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                </div>
                <h3 className="text-white text-lg font-medium">Withdrawal Request Submitted</h3>
                <p className="text-gray-400 text-sm mt-2">
                    Your withdrawal request is pending admin approval. Check the History tab for updates.
                </p>
            </div>
        );
    }

    return (
        <form onSubmit={handleWithdraw} className="space-y-6">
            {/* Balance Display */}
            <div className="bg-gray-700 rounded-lg p-4 text-center">
                <p className="text-gray-400 text-sm">Available Balance</p>
                <p className="text-white text-2xl font-bold">{userCredit} Credits</p>
            </div>

            {/* Chain Selector */}
            <div>
                <label className="block text-gray-400 text-sm mb-2">Network</label>
                <ChainSelector
                    selectedChain={selectedChain}
                    onSelect={setSelectedChain}
                    disabled={loading}
                />
            </div>

            {/* Token Selector */}
            <div>
                <label className="block text-gray-400 text-sm mb-2">Token</label>
                <TokenSelector
                    selectedToken={selectedToken}
                    onSelect={setSelectedToken}
                    disabled={loading}
                />
            </div>

            {/* Amount Input */}
            <div>
                <label className="block text-gray-400 text-sm mb-2">Amount</label>
                <div className="relative">
                    <input
                        type="number"
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                        placeholder="Enter amount"
                        min="10"
                        max={userCredit}
                        disabled={loading}
                        className="w-full bg-gray-700 text-white rounded-lg px-4 py-3 pr-16
                                   border border-gray-600 focus:border-blue-500 focus:outline-none
                                   disabled:opacity-50"
                        required
                    />
                    <button
                        type="button"
                        onClick={handleMaxAmount}
                        disabled={loading}
                        className="absolute right-2 top-1/2 -translate-y-1/2
                                   bg-blue-600 hover:bg-blue-700 disabled:opacity-50
                                   text-white text-xs px-3 py-1 rounded"
                    >
                        MAX
                    </button>
                </div>
                <p className="text-gray-500 text-xs mt-1">Minimum: 10 credits</p>
            </div>

            {/* Destination Address */}
            <div>
                <label className="block text-gray-400 text-sm mb-2">Destination Address</label>
                <input
                    type="text"
                    value={toAddress}
                    onChange={(e) => setToAddress(e.target.value)}
                    placeholder="0x..."
                    disabled={loading}
                    className={`w-full bg-gray-700 text-white rounded-lg px-4 py-3
                               border focus:outline-none font-mono text-sm
                               disabled:opacity-50
                               ${toAddress && !isValidAddress(toAddress)
                                   ? 'border-red-500 focus:border-red-500'
                                   : 'border-gray-600 focus:border-blue-500'}`}
                    required
                />
                {toAddress && !isValidAddress(toAddress) && (
                    <p className="text-red-400 text-xs mt-1">Invalid address format</p>
                )}
            </div>

            {/* Error Display */}
            {error && (
                <div className="bg-red-900/30 border border-red-600/50 rounded-lg p-3">
                    <p className="text-red-400 text-sm">{error}</p>
                </div>
            )}

            {/* Submit Button */}
            <button
                type="submit"
                disabled={!canSubmit}
                className="w-full bg-red-600 hover:bg-red-700 disabled:bg-gray-600
                           disabled:cursor-not-allowed text-white font-medium py-3
                           rounded-lg transition-colors"
            >
                {loading ? 'Processing...' : 'Request Withdrawal'}
            </button>

            {/* Fee Info */}
            <p className="text-gray-500 text-xs text-center">
                Withdrawals require admin approval. Network fees will be deducted from the withdrawal amount.
            </p>
        </form>
    );
};

export default WithdrawTab;
