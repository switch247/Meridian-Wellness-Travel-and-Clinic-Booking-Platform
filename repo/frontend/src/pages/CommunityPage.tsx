import { Accordion, AccordionDetails, AccordionSummary, Alert, Avatar, Box, Button, Divider, Paper, Stack, TextField, Typography } from '@mui/material';
import ExpandMoreRoundedIcon from '@mui/icons-material/ExpandMoreRounded';
import { useEffect, useMemo, useState } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { SectionHeader } from '../components/common/SectionHeader';

export function CommunityPage() {
  const { token } = useAuth();
  const [posts, setPosts] = useState<Array<Record<string, unknown>>>([]);
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [activePost, setActivePost] = useState<number | null>(null);
  const [commentBody, setCommentBody] = useState('');
  const [comments, setComments] = useState<Record<number, Array<Record<string, unknown>>>>({});
  const [notification, setNotification] = useState<string | null>(null);

  const emptyState = useMemo(() => posts.length === 0, [posts]);

  async function loadPosts() {
    if (!token) return;
    const out = await api.communityPosts(token);
    setPosts(out.items || []);
  }

  useEffect(() => {
    loadPosts().catch(() => {});
  }, [token]);

  const loadComments = async (postId: number) => {
    if (!token) return;
    const out = await api.communityComments(token, postId);
    setComments((prev) => ({ ...prev, [postId]: out.items || [] }));
  };

  const handleAction = async (action: () => Promise<void>, success?: string) => {
    if (!token) return;
    try {
      await action();
      if (success) {
        setNotification(success);
      }
    } catch (err) {
      setError((err as Error).message);
    }
  };

  return (
    <Stack spacing={2.5}>
      <SectionHeader title="Community" subtitle="Threaded travel Q&A, moderation, and social signals." />

      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6" sx={{ mb: 1 }}>Create a Discussion</Typography>
        {error && <Alert severity="error" sx={{ mb: 1.5 }}>{error}</Alert>}
        {notification && (
          <Alert severity="success" sx={{ mb: 1.5 }} onClose={() => setNotification(null)}>
            {notification}
          </Alert>
        )}
        <Stack spacing={1.5}>
          <TextField label="Title" value={title} onChange={(e) => setTitle(e.target.value)} />
          <TextField label="Body" multiline minRows={3} value={body} onChange={(e) => setBody(e.target.value)} />
          <Button
            variant="contained"
            onClick={async () => {
              if (!token) return;
              await api.createCommunityPost(token, { title, body });
              setTitle('');
              setBody('');
              await loadPosts();
            }}
          >
            Publish Post
          </Button>
        </Stack>
      </Paper>

      <Paper sx={{ p: 2.5 }}>
        <Typography variant="h6">Recent Threads</Typography>
        {emptyState ? (
          <Alert severity="info" sx={{ mt: 1.5 }}>
            The community is quiet. Start a discussion about travel preferences or provider experiences.
          </Alert>
        ) : (
          <Stack spacing={1.5} sx={{ mt: 1.2 }}>
            {posts.map((post) => {
              const postId = Number(post.id);
              const postAuthor = String(post.authorUserId || 'unknown');
              const destination = post.destinationId ? `Destination #${post.destinationId}` : 'General';
              return (
                <Accordion key={postId} variant="outlined" expanded={activePost === postId} onChange={(_, expanded) => {
                  setActivePost(expanded ? postId : null);
                  if (expanded) {
                    loadComments(postId);
                  }
                }}>
                  <AccordionSummary expandIcon={<ExpandMoreRoundedIcon />}>
                    <Stack direction="row" alignItems="center" spacing={1} sx={{ flexGrow: 1 }}>
                      <Avatar sx={{ bgcolor: '#0d6e6e' }}>{String(postAuthor)[0]?.toUpperCase() || 'U'}</Avatar>
                      <Box>
                        <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>{String(post.title)}</Typography>
                        <Typography variant="body2" color="text.secondary">
                          {destination} · {String(post.status || 'active')}
                        </Typography>
                      </Box>
                    </Stack>
                    <Chip label={`Author ${postAuthor}`} size="small" />
                  </AccordionSummary>
                  <AccordionDetails>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 1.2 }}>
                      {String(post.body || '')}
                    </Typography>
                    <Stack direction="row" spacing={1} flexWrap="wrap">
                      <Button size="small" variant="outlined" onClick={() => handleAction(async () => {
                        if (!token) return;
                        await api.likeTarget(token, { targetType: 'post', targetId: postId });
                      }, 'Post liked')}>
                        Like
                      </Button>
                      <Button size="small" variant="outlined" onClick={() => {
                        const packageId = Number(post.destinationId ?? 0);
                        if (!packageId) return;
                        handleAction(async () => {
                          if (!token) return;
                          await api.favoritePackage(token, { packageId });
                        }, 'Package favorited');
                      }}>
                        Favorite Package
                      </Button>
                      <Button size="small" variant="outlined" onClick={() => {
                        const targetId = Number(post.authorUserId ?? 0);
                        if (!targetId) return;
                        handleAction(async () => {
                          if (!token) return;
                          await api.followUser(token, { userId: targetId });
                        }, 'Follow request sent');
                      }}>
                       Follow Author
                     </Button>
                      <Button size="small" color="warning" variant="outlined" onClick={() => {
                        const targetId = Number(post.authorUserId ?? 0);
                        if (!targetId) return;
                        handleAction(async () => {
                          if (!token) return;
                          await api.blockUser(token, { userId: targetId });
                        }, 'User blocked');
                      }}>
                       Block Author
                     </Button>
                      <Button size="small" color="error" variant="outlined" onClick={() => handleAction(async () => {
                        if (!token) return;
                        await api.reportContent(token, { targetType: 'post', targetId: postId, reason: 'Requires moderation' });
                      }, 'Report submitted')}>
                        Report
                      </Button>
                    </Stack>
                    <Divider sx={{ my: 1.2 }} />
                    {activePost === postId && (
                      <Stack spacing={1.2}>
                        {(comments[postId] || []).map((comment) => (
                          <Paper key={String(comment.id)} variant="outlined" sx={{ p: 1.2 }}>
                            <Typography variant="body2">{String(comment.body)}</Typography>
                            <Button size="small" variant="text" onClick={() => handleAction(async () => {
                              if (!token) return;
                              await api.likeTarget(token, { targetType: 'comment', targetId: Number(comment.id) });
                            }, 'Comment liked')}>
                              Like
                            </Button>
                          </Paper>
                        ))}
                        <Stack direction="row" spacing={1}>
                          <TextField
                            label="Reply"
                            fullWidth
                            value={commentBody}
                            onChange={(e) => setCommentBody(e.target.value)}
                          />
                          <Button variant="contained" onClick={async () => {
                            if (!token || !commentBody) return;
                            await api.addComment(token, postId, { body: commentBody });
                            setCommentBody('');
                            await loadComments(postId);
                          }}>
                            Reply
                          </Button>
                        </Stack>
                      </Stack>
                    )}
                  </AccordionDetails>
                </Accordion>
              );
            })}
          </Stack>
        )}
      </Paper>
    </Stack>
  );
}
