import { Alert, Button, Paper, Stack, TextField, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { SectionHeader } from '../components/common/SectionHeader';

export function CommunityPage() {
  const { token } = useAuth();
  const [items, setItems] = useState<Array<Record<string, unknown>>>([]);
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [activePost, setActivePost] = useState<number | null>(null);
  const [comments, setComments] = useState<Array<Record<string, unknown>>>([]);
  const [commentBody, setCommentBody] = useState('');
  const [reportReason, setReportReason] = useState('Inappropriate content');
  const [targetUserId, setTargetUserId] = useState('1');
  const [packageId, setPackageId] = useState('1');

  async function load() {
    if (!token) return;
    const out = await api.communityPosts(token);
    setItems(out.items || []);
  }

  useEffect(() => {
    load().catch(() => {});
  }, [token]);

  async function loadComments(postId: number) {
    if (!token) return;
    const out = await api.communityComments(token, postId);
    setComments(out.items || []);
  }

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Community" subtitle="Q&A and threaded travel/provider discussion." />
      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6" sx={{ mb: 1.5 }}>Create Post</Typography>
        {error && <Alert severity="error" sx={{ mb: 1.5 }}>{error}</Alert>}
        <Stack spacing={1.5}>
          <TextField label="Title" value={title} onChange={(e) => setTitle(e.target.value)} />
          <TextField label="Body" multiline minRows={3} value={body} onChange={(e) => setBody(e.target.value)} />
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.2}>
            <TextField label="Target User ID (follow/block)" value={targetUserId} onChange={(e) => setTargetUserId(e.target.value)} />
            <TextField label="Package ID (favorite)" value={packageId} onChange={(e) => setPackageId(e.target.value)} />
          </Stack>
          <Button variant="contained" onClick={async () => {
            try {
              if (!token) return;
              await api.createCommunityPost(token, { title, body });
              setTitle('');
              setBody('');
              await load();
            } catch (e) {
              setError((e as Error).message);
            }
          }}>Publish</Button>
        </Stack>
      </Paper>
      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Recent Posts</Typography>
        {items.length === 0 ? (
          <Alert severity="info" sx={{ mt: 1.5 }}>No community posts yet. Start the first conversation.</Alert>
        ) : (
          <Stack spacing={1.2} sx={{ mt: 1.5 }}>
            {items.map((p, i) => (
              <Paper key={i} variant="outlined" sx={{ p: 1.5 }}>
                <Typography variant="subtitle1">{String(p.title)}</Typography>
                <Typography variant="body2" color="text.secondary">{String(p.body)}</Typography>
                <Stack direction="row" spacing={1} sx={{ mt: 1 }} flexWrap="wrap">
                  <Button size="small" variant="outlined" onClick={async () => {
                    if (!token) return;
                    setActivePost(Number(p.id));
                    await loadComments(Number(p.id));
                  }}>Thread</Button>
                  <Button size="small" variant="outlined" onClick={async () => {
                    if (!token) return;
                    await api.favoritePackage(token, { packageId: Number(packageId) });
                  }}>Favorite</Button>
                  <Button size="small" variant="outlined" onClick={async () => {
                    if (!token) return;
                    await api.followUser(token, { userId: Number(targetUserId) });
                  }}>Follow</Button>
                  <Button size="small" color="warning" variant="outlined" onClick={async () => {
                    if (!token) return;
                    await api.blockUser(token, { userId: Number(targetUserId) });
                  }}>Block</Button>
                  <Button size="small" color="error" variant="outlined" onClick={async () => {
                    if (!token) return;
                    await api.reportContent(token, { targetType: 'post', targetId: Number(p.id), reason: reportReason });
                  }}>Report</Button>
                  <Button size="small" variant="outlined" onClick={async () => {
                  if (!token) return;
                  await api.likeTarget(token, { targetType: 'post', targetId: Number(p.id) });
                  }}>Like</Button>
                </Stack>
              </Paper>
            ))}
          </Stack>
        )}
      </Paper>
      {activePost && (
        <Paper sx={{ p: 2.5 }}>
          <Typography variant="h6">Thread #{activePost}</Typography>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.2} sx={{ mt: 1 }}>
            <TextField label="Report reason" fullWidth value={reportReason} onChange={(e) => setReportReason(e.target.value)} />
            <Button variant="outlined" onClick={async () => {
              if (!token) return;
              await api.reportContent(token, { targetType: 'post', targetId: activePost, reason: reportReason });
            }}>Report Post</Button>
          </Stack>
          <Stack spacing={1.2} sx={{ mt: 1.5 }}>
            {comments.length === 0 ? <Alert severity="info">No comments yet.</Alert> : comments.map((c, idx) => (
              <Paper key={idx} variant="outlined" sx={{ p: 1.2 }}>
                <Typography variant="body2">{String(c.body)}</Typography>
                <Button size="small" onClick={async () => {
                  if (!token) return;
                  await api.likeTarget(token, { targetType: 'comment', targetId: Number(c.id) });
                }}>Like comment</Button>
              </Paper>
            ))}
          </Stack>
          <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.2} sx={{ mt: 1.5 }}>
            <TextField label="Reply" fullWidth value={commentBody} onChange={(e) => setCommentBody(e.target.value)} />
            <Button variant="contained" onClick={async () => {
              if (!token || !commentBody) return;
              await api.addComment(token, activePost, { body: commentBody });
              setCommentBody('');
              await loadComments(activePost);
            }}>Reply</Button>
          </Stack>
        </Paper>
      )}
    </Stack>
  );
}
