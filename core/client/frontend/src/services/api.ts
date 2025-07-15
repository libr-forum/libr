import { SendInput,FetchAll } from "../../wailsjs/go/main/App"; 
import axios from 'axios';
import { Community, Message, User } from '../store/useAppStore';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
});

// Mock data for demo purposes
const mockCommunities: Community[] = [
  {
    id: '1',
    name: 'cryptography',
    topic: 'Discussion about cryptographic protocols and implementations',
    memberCount: 1247,
    isJoined: true,
    requiresApproval: false,
  },
  {
    id: '2',
    name: 'privacy-tech',
    topic: 'Privacy technologies and digital rights',
    memberCount: 892,
    isJoined: true,
    requiresApproval: false,
  },
  {
    id: '3',
    name: 'decentralized-web',
    topic: 'Building the decentralized internet',
    memberCount: 567,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '4',
    name: 'blockchain-dev',
    topic: 'Blockchain development and smart contracts',
    memberCount: 2341,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '5',
    name: 'web3-gaming',
    topic: 'Web3 gaming and NFT discussions',
    memberCount: 876,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '6',
    name: 'defi-protocols',
    topic: 'Decentralized Finance protocols and analysis',
    memberCount: 1654,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '7',
    name: 'zero-knowledge',
    topic: 'Zero-knowledge proofs and privacy solutions',
    memberCount: 432,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '8',
    name: 'dao-governance',
    topic: 'DAO governance and tokenomics',
    memberCount: 987,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '9',
    name: 'layer2-scaling',
    topic: 'Layer 2 scaling solutions and optimizations',
    memberCount: 723,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '10',
    name: 'censorship-resistance',
    topic: 'Building censorship-resistant systems',
    memberCount: 1123,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '11',
    name: 'peer-to-peer',
    topic: 'P2P networking and distributed systems',
    memberCount: 645,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '12',
    name: 'digital-identity',
    topic: 'Self-sovereign identity and digital credentials',
    memberCount: 498,
    isJoined: false,
    requiresApproval: false,
  },
];

const mockMessages: Message[] = [
  {
    id: '1',
    content: 'Has anyone tried implementing zero-knowledge proofs with the new Circom 2.0?',
    authorId: 'user123',
    authorAlias: 'CryptoExplorer',
    timestamp: new Date(Date.now() - 1000 * 60 * 5),
    communityId: '1',
    status: 'approved',
  },
  {
    id: '2',
    content: 'The latest paper on post-quantum cryptography is fascinating. Here\'s the link: https://eprint.iacr.org/...',
    authorId: 'user456',
    authorAlias: 'QuantumResearcher',
    timestamp: new Date(Date.now() - 1000 * 60 * 15),
    communityId: '1',
    status: 'approved',
  },
  {
    id: '3',
    content: 'Looking for collaborators on a new privacy-preserving protocol. DM me if interested!',
    authorId: 'user789',
    authorAlias: 'PrivacyAdvocate',
    timestamp: new Date(Date.now() - 1000 * 60 * 30),
    communityId: '1',
    status: 'pending',
  },
];

export const apiService = {
  // Auth
  async authenticate(publicKey: string): Promise<User> {
    // Mock authentication
    return {
      id: 'user123',
      publicKey,
      alias: 'CryptoExplorer',
      role: 'moderator', // Demo as moderator
    };
  },

  // Relay connect


  // Communities
  async getCommunities(): Promise<Community[]> {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500));
    return mockCommunities;
  },

  async joinCommunity(communityId: string): Promise<void> {
    await new Promise(resolve => setTimeout(resolve, 300));
    console.log(`Joined community: ${communityId}`);
  },

  // Messages
    async getMessages(_communityId: string): Promise<Message[]> {
    try {
      const response = await FetchAll(); // returns []string
      // Convert each string into a Message object
      return response.map((line): Message => {
        // Expected format: "Sender: <sender> | Msg: <msg> | Time: <timestamp>"
        const senderMatch = line.match(/Sender: (.*?) \|/);
        const msgMatch = line.match(/Msg: (.*?) \|/);
        const timeMatch = line.match(/Time: (\d+)/);

        return {
          id: timeMatch?.[1] || Date.now().toString(),
          content: msgMatch?.[1] || "",
          authorId: senderMatch?.[1] || "unknown",
          authorAlias: senderMatch?.[1] || "unknown",
          timestamp: new Date(parseInt(timeMatch?.[1] || "0") * 1000),
          communityId: _communityId,
          status: "approved",
        };
      });
    } catch (err) {
      console.error("Failed to fetch messages:", err);
      return [];
    }
  },


  async sendMessage(communityId: string, content: string): Promise<Message> {
    const result = await SendInput(content);

    const approved = result.includes("✅");
    const newMessage: Message = {
      id: Date.now().toString(),
      content,
      authorId: 'user123',
      authorAlias: 'CryptoExplorer',
      timestamp: new Date(),
      communityId,
      status: approved ? 'approved' : 'rejected', // timeout = rejected
    };

    return newMessage;
  },


  // Moderation
  async getModerationLogs(): Promise<Message[]> {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockMessages.filter(m => m.status !== 'approved');
  },

  async moderateMessage(messageId: string, action: 'approve' | 'reject', note?: string): Promise<void> {
    await new Promise(resolve => setTimeout(resolve, 300));
    console.log(`Message ${messageId} ${action}ed with note: ${note}`);
  },
};
