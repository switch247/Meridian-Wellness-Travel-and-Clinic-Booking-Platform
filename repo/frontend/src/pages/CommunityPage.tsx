import {
  Box,
  Card,
  CardContent,
  Typography,
  Avatar,
  Stack,
  Button,
  TextField,
  Divider,
  IconButton,
  Chip,
  Paper,
  Snackbar,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import ThumbUpIcon from '@mui/icons-material/ThumbUp';
import SendIcon from '@mui/icons-material/Send';
import CloseIcon from '@mui/icons-material/Close';
import { useEffect, useState, useMemo } from 'react';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';

export function CommunityPage() {
  const { token, me } = useAuth();

  const [posts, setPosts] = useState<any[]>([]);
  const [selectedPost, setSelectedPost] = useState<any>(null);
  const [comments, setComments] = useState<Record<number, any[]>>({});

  const [newPostTitle, setNewPostTitle] = useState('');
  const [newPostBody, setNewPostBody] = useState('');
  const [newPostDestination, setNewPostDestination] = useState('');
  const [commentBody, setCommentBody] = useState('');

  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [destinationOptions, setDestinationOptions] = useState<Array<{ id: number; label: string }>>([]);

  const isEmpty = useMemo(() => posts.length === 0, [posts]);

  const loadPosts = async () => {
    if (!token) return;
    try {
      const res = await api.communityPosts(token);
      setPosts(res.items || []);
    } catch (err) {
      setError('Failed to load discussions');
    }
  };

  const loadComments = async (postId: number) => {
    if (!token) return;
    try {
      const res = await api.communityComments(token, postId);
      setComments((prev) => ({ ...prev, [postId]: res.items || [] }));
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    loadPosts();

    // Build destination options from backend catalog master data (no hardcoded IDs).
    api.routes()
      .then((res) => {
        const seen = new Set<number>();
        const opts: Array<{ id: number; label: string }> = [];
        (res.items || []).forEach((it) => {
          const id = Number(it.destinationId || 0);
          if (id > 0 && !seen.has(id)) {
            seen.add(id);
            opts.push({ id, label: `Destination #${id}` });
          }
        });
        setDestinationOptions(opts.sort((a, b) => a.id - b.id));
      })
      .catch(() => {
        setDestinationOptions([]);
      });
  }, [token]);

  const destinationLabelById = useMemo(() => {
    const map: Record<number, string> = {};
    destinationOptions.forEach((d) => {
      map[d.id] = d.label;
    });
    return map;
  }, [destinationOptions]);

  const handleSelectPost = async (post: any) => {
    setSelectedPost(post);
    await loadComments(Number(post.id));
  };

  const handleCreatePost = async () => {
    if (!token || !newPostTitle.trim() || !newPostBody.trim()) return;
    setLoading(true);
    try {
      await api.createCommunityPost(token, {
        title: newPostTitle,
        body: newPostBody,
        destinationId: newPostDestination ? Number(newPostDestination) : undefined,
      });
      setNewPostTitle('');
      setNewPostBody('');
      setNewPostDestination('');
      setCreateDialogOpen(false);
      setSuccess('Discussion created');
      await loadPosts();
    } catch (err) {
      setError('Failed to create discussion');
    } finally {
      setLoading(false);
    }
  };

  const handleAddComment = async () => {
    if (!token || !commentBody.trim() || !selectedPost) return;
    try {
      await api.addComment(token, Number(selectedPost.id), { body: commentBody });
      setCommentBody('');
      await loadComments(Number(selectedPost.id));
      setSuccess('Reply posted');
    } catch (err) {
      setError('Failed to post reply');
    }
  };

  const handleLike = async (targetType: 'post' | 'comment', targetId: number) => {
    if (!token) return;
    try {
      await api.likeTarget(token, { targetType, targetId });
      setSuccess('Liked');
      if (targetType === 'post' && selectedPost) {
        await loadComments(Number(selectedPost.id));
      }
    } catch (err) {
      setError('Failed to like');
    }
  };

  return (
    <Box sx={{ height: '100vh', display: 'flex', flexDirection: 'column', bgcolor: 'grey.50' }}>
      {/* Top Header */}
      <Box sx={{ p: 3, borderBottom: '1px solid', borderColor: 'divider', bgcolor: 'background.paper' }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Typography variant="h5" fontWeight={600}>
            Community Discussions
          </Typography>

          <Button
            variant="outlined"
            startIcon={<AddIcon />}
            onClick={() => setCreateDialogOpen(true)}
            sx={{ borderRadius: 1 }}
          >
            New Discussion
          </Button>
        </Stack>
      </Box>

      <Box sx={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
        {/* Left: Conversation List */}
        <Box
          sx={{
            width: { xs: '100%', md: 360 },
            borderRight: '1px solid',
            borderColor: 'divider',
            bgcolor: 'background.paper',
            overflowY: 'auto',
            p: 3,
          }}
        >
          <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 2, px: 1 }}>
            All Discussions
          </Typography>

          {isEmpty ? (
            <Box sx={{ textAlign: 'center', py: 10 }}>
              <Typography variant="h6" color="text.secondary" gutterBottom>
                No discussions yet
              </Typography>
              <Button
                variant="outlined"
                startIcon={<AddIcon />}
                onClick={() => setCreateDialogOpen(true)}
                sx={{ mt: 2, borderRadius: 1 }}
              >
                Start the first discussion
              </Button>
            </Box>
          ) : (
            <Stack spacing={1}>
              {posts.map((post) => {
                const postId = Number(post.id);
                const isSelected = selectedPost?.id === postId;
                const commentCount = comments[postId]?.length || 0;

                return (
                  <Card
                    key={postId}
                    onClick={() => handleSelectPost(post)}
                    sx={{
                      cursor: 'pointer',
                      borderRadius: 1,           // reduced rounding
                      border: '1px solid',
                      borderColor: isSelected ? 'primary.main' : 'divider',
                      bgcolor: isSelected ? 'action.selected' : 'background.paper',
                      '&:hover': {
                        borderColor: 'primary.light',
                        bgcolor: 'action.hover',
                      },
                    }}
                  >
                    <CardContent sx={{ p: 2 }}>
                      <Stack direction="row" spacing={2} alignItems="flex-start">
                        <Avatar sx={{ width: 34, height: 34, bgcolor: 'primary.main', fontSize: '0.9rem' }}>
                          {String(post.authorUserId || 'U')[0]?.toUpperCase()}
                        </Avatar>

                        <Box sx={{ flex: 1, minWidth: 0 }}>
                          <Typography variant="subtitle2" noWrap>
                            User {post.authorUserId}
                          </Typography>

                          <Typography
                            variant="body2"
                            sx={{
                              mt: 0.5,
                              mb: 0.5,
                              fontWeight: 600,
                            }}
                            noWrap
                          >
                            {String(post.title || '')}
                          </Typography>

                          <Typography
                            variant="body2"
                            sx={{
                              display: '-webkit-box',
                              WebkitLineClamp: 2,
                              WebkitBoxOrient: 'vertical',
                              overflow: 'hidden',
                              color: 'text.secondary',
                            }}
                          >
                            {String(post.body || '')}
                          </Typography>

                          <Stack direction="row" spacing={1.5} alignItems="center">
                            <Chip
                              label={post.destinationId ? (destinationLabelById[Number(post.destinationId)] || `Destination #${post.destinationId}`) : 'General'}
                              size="small"
                              variant="outlined"
                              sx={{ height: 22, fontSize: '0.75rem', borderRadius: 1 }}
                            />
                            <Typography variant="caption" color="text.secondary">
                              {commentCount} replies
                            </Typography>
                          </Stack>
                        </Box>
                      </Stack>
                    </CardContent>
                  </Card>
                );
              })}
            </Stack>
          )}
        </Box>

        {/* Right: Thread View */}
        <Box
          sx={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
            bgcolor: 'grey.50',
          }}
        >
          {!selectedPost ? (
            <Box
              sx={{
                flex: 1,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                textAlign: 'center',
                p: 4,
              }}
            >
              <Box>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  Select a discussion
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Click on a conversation from the left to view and reply
                </Typography>
              </Box>
            </Box>
          ) : (
            <>
              {/* Thread Header */}
              <Box sx={{ p: 3, bgcolor: 'background.paper', borderBottom: '1px solid', borderColor: 'divider' }}>
                <Stack direction="row" spacing={2} alignItems="center">
                  <Avatar sx={{ bgcolor: 'primary.main' }}>
                    {String(selectedPost.authorUserId || 'U')[0]?.toUpperCase()}
                  </Avatar>
                  <Box sx={{ flex: 1 }}>
                    <Typography variant="subtitle1" fontWeight={600}>
                      User {selectedPost.authorUserId}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {selectedPost.destinationId ? (destinationLabelById[Number(selectedPost.destinationId)] || `Destination #${selectedPost.destinationId}`) : 'General'}
                    </Typography>
                  </Box>
                </Stack>

                <Typography variant="h6" sx={{ mt: 2, mb: 1, fontWeight: 600 }}>
                  {selectedPost.title}
                </Typography>

                <Typography variant="body1" sx={{ mt: 1, lineHeight: 1.7 }}>
                  {selectedPost.body}
                </Typography>

                <Stack direction="row" spacing={1} sx={{ mt: 2 }}>
                  <IconButton size="small" onClick={() => handleLike('post', Number(selectedPost.id))}>
                    <ThumbUpIcon fontSize="small" />
                  </IconButton>
                </Stack>
              </Box>

              {/* Comments */}
              <Box sx={{ flex: 1, p: 3, overflowY: 'auto' }}>
                <Stack spacing={2.5}>
                  {(comments[Number(selectedPost.id)] || []).map((comment: any) => (
                    <Paper key={comment.id} sx={{ p: 3, borderRadius: 1 }}>
                      <Stack direction="row" spacing={2}>
                        <Avatar sx={{ width: 32, height: 32, bgcolor: 'grey.400' }}>
                          {String(comment.authorUserId || 'U')[0]?.toUpperCase()}
                        </Avatar>
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="subtitle2">
                            User {comment.authorUserId}
                          </Typography>
                          <Typography variant="body2" sx={{ mt: 0.5, lineHeight: 1.6 }}>
                            {comment.body}
                          </Typography>

                          <IconButton
                            size="small"
                            sx={{ mt: 1 }}
                            onClick={() => handleLike('comment', Number(comment.id))}
                          >
                            <ThumbUpIcon fontSize="small" />
                          </IconButton>
                        </Box>
                      </Stack>
                    </Paper>
                  ))}
                </Stack>
              </Box>

              {/* Reply Input */}
              <Box sx={{ p: 3, bgcolor: 'background.paper', borderTop: '1px solid', borderColor: 'divider' }}>
                <Stack direction="row" spacing={2} alignItems="flex-end">
                  <Avatar sx={{ bgcolor: 'primary.main' }}>
                    {me?.username?.[0]?.toUpperCase() || 'U'}
                  </Avatar>

                  <TextField
                    fullWidth
                    placeholder="Write a reply..."
                    value={commentBody}
                    onChange={(e) => setCommentBody(e.target.value)}
                    multiline
                    maxRows={4}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault();
                        handleAddComment();
                      }
                    }}
                    sx={{ bgcolor: 'background.paper' }}
                  />

                  <IconButton
                    color="primary"
                    onClick={handleAddComment}
                    disabled={!commentBody.trim()}
                  >
                    <SendIcon />
                  </IconButton>
                </Stack>
              </Box>
            </>
          )}
        </Box>
      </Box>

      {/* Create Discussion Dialog */}
      <Dialog
        open={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>
          New Discussion
          <IconButton
            onClick={() => setCreateDialogOpen(false)}
            sx={{ position: 'absolute', right: 8, top: 8 }}
          >
            <CloseIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label="Title"
              placeholder="Enter discussion title..."
              value={newPostTitle}
              onChange={(e) => setNewPostTitle(e.target.value)}
            />

            <TextField
              fullWidth
              label="Description"
              placeholder="Share your experience or ask a question..."
              value={newPostBody}
              onChange={(e) => setNewPostBody(e.target.value)}
              multiline
              rows={4}
            />

            <FormControl fullWidth>
              <InputLabel>Category (optional)</InputLabel>
              <Select
                value={newPostDestination}
                onChange={(e) => setNewPostDestination(e.target.value)}
                label="Category (optional)"
              >
                <MenuItem value="">General Discussion</MenuItem>
                {destinationOptions.map((d) => (
                  <MenuItem key={d.id} value={String(d.id)}>{d.label}</MenuItem>
                ))}
              </Select>
            </FormControl>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateDialogOpen(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={handleCreatePost}
            disabled={!newPostTitle.trim() || !newPostBody.trim() || loading}
          >
            {loading ? 'Posting...' : 'Post Discussion'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Notifications */}
      <Snackbar open={!!error} autoHideDuration={5000} onClose={() => setError(null)}>
        <Alert severity="error" onClose={() => setError(null)}>{error}</Alert>
      </Snackbar>

      <Snackbar open={!!success} autoHideDuration={4000} onClose={() => setSuccess(null)}>
        <Alert severity="success" onClose={() => setSuccess(null)}>{success}</Alert>
      </Snackbar>
    </Box>
  );
}