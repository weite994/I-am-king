import 'package:flutter/material.dart';
import 'package:video_player/video_player.dart';
import 'package:image_picker/image_picker.dart';
import 'dart:io';

void main() {
  runApp(TikTakApp());
}

class TikTakApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'تيك تك - همدان مشهل',
      theme: ThemeData(
        primarySwatch: Colors.orange,
        fontFamily: 'Arial',
      ),
      home: LoginPage(),
      debugShowCheckedModeBanner: false,
    );
  }
}

// --- صفحة تسجيل دخول وهمي ---
class LoginPage extends StatelessWidget {
  final TextEditingController userController = TextEditingController();

  void login(BuildContext context) {
    if (userController.text.isNotEmpty) {
      Navigator.pushReplacement(
          context,
          MaterialPageRoute(
              builder: (_) => HomePage(username: userController.text)));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('تسجيل الدخول')),
      body: Padding(
        padding: EdgeInsets.all(20),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            TextField(
              controller: userController,
              decoration: InputDecoration(
                  labelText: 'اسم المستخدم', border: OutlineInputBorder()),
            ),
            SizedBox(height: 20),
            ElevatedButton(
              onPressed: () => login(context),
              child: Text('دخول'),
            ),
          ],
        ),
      ),
    );
  }
}

// --- الصفحة الرئيسية: قائمة فيديوهات ---
class HomePage extends StatefulWidget {
  final String username;
  HomePage({required this.username});

  @override
  _HomePageState createState() => _HomePageState();
}

class VideoItem {
  final String url;
  bool liked;
  int likes;
  List<String> comments;

  VideoItem(
      {required this.url,
      this.liked = false,
      this.likes = 0,
      List<String>? comments})
      : comments = comments ?? [];
}

class _HomePageState extends State<HomePage> {
  List<VideoItem> videos = [
    VideoItem(
        url:
            'https://flutter.github.io/assets-for-api-docs/assets/videos/butterfly.mp4',
        likes: 10,
        comments: ['جميل!', 'روعة']),
    VideoItem(
        url:
            'https://flutter.github.io/assets-for-api-docs/assets/videos/bee.mp4',
        likes: 5,
        comments: ['حلو!', 'ممتاز']),
  ];

  int currentIndex = 0;
  late VideoPlayerController _controller;

  @override
  void initState() {
    super.initState();
    _loadVideo(currentIndex);
  }

  void _loadVideo(int index) {
    if (_controller != null) {
      _controller.pause();
      _controller.dispose();
    }
    _controller = VideoPlayerController.network(videos[index].url)
      ..initialize().then((_) {
        setState(() {});
        _controller.play();
        _controller.setLooping(true);
      });
  }

  void nextVideo() {
    setState(() {
      currentIndex = (currentIndex + 1) % videos.length;
      _loadVideo(currentIndex);
    });
  }

  void previousVideo() {
    setState(() {
      currentIndex = (currentIndex - 1 + videos.length) % videos.length;
      _loadVideo(currentIndex);
    });
  }

  void toggleLike() {
    setState(() {
      videos[currentIndex].liked = !videos[currentIndex].liked;
      if (videos[currentIndex].liked) {
        videos[currentIndex].likes += 1;
      } else {
        videos[currentIndex].likes -= 1;
      }
    });
  }

  void addComment(String comment) {
    setState(() {
      videos[currentIndex].comments.add(comment);
    });
  }

  // رفع فيديو (اختيار فيديو من المعرض فقط)
  Future<void> pickVideo() async {
    final picker = ImagePicker();
    final pickedFile = await picker.pickVideo(source: ImageSource.gallery);
    if (pickedFile != null) {
      setState(() {
        videos.add(VideoItem(url: pickedFile.path, likes: 0));
        currentIndex = videos.length - 1;
        _loadVideo(currentIndex);
      });
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  TextEditingController commentController = TextEditingController();

  @override
  Widget build(BuildContext context) {
    final video = videos[currentIndex];
    return Scaffold(
      appBar: AppBar(
        title: Text('تيك تك - همدان مشهل'),
        actions: [
          IconButton(
              icon: Icon(Icons.add),
              onPressed: () async {
                await pickVideo();
              }),
          IconButton(
              icon: Icon(Icons.person),
              onPressed: () {
                Navigator.push(
                    context,
                    MaterialPageRoute(
                        builder: (_) =>
                            ProfilePage(username: widget.username)));
              }),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: _controller.value.isInitialized
                ? AspectRatio(
                    aspectRatio: _controller.value.aspectRatio,
                    child: VideoPlayer(_controller),
                  )
                : Center(child: CircularProgressIndicator()),
          ),
          Padding(
            padding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Row(
              children: [
                IconButton(
                  icon: Icon(
                    video.liked ? Icons.favorite : Icons.favorite_border,
                    color: video.liked ? Colors.red : Colors.black,
                  ),
                  onPressed: toggleLike,
                ),
                Text('${video.likes}'),
                Spacer(),
                IconButton(
                  icon: Icon(Icons.skip_previous),
                  onPressed: previousVideo,
                ),
                IconButton(
                  icon: Icon(Icons.skip_next),
                  onPressed: nextVideo,
                ),
              ],
            ),
          ),
          // التعليقات
          Expanded(
            child: ListView.builder(
              itemCount: video.comments.length,
              itemBuilder: (context, index) {
                return ListTile(
                  leading: Icon(Icons.comment),
                  title: Text(video.comments[index]),
                );
              },
            ),
          ),
          Padding(
            padding: EdgeInsets.symmetric(horizontal: 12),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: commentController,
                    decoration: InputDecoration(
                        hintText: 'أضف تعليق...',
                        border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12))),
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.send),
                  onPressed: () {
                    if (commentController.text.trim().isNotEmpty) {
                      addComment(commentController.text.trim());
                      commentController.clear();
                    }
                  },
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// --- صفحة الملف الشخصي ---
class ProfilePage extends StatelessWidget {
  final String username;
  ProfilePage({required this.username});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('الملف الشخصي'),
      ),
      body: Center(
        child: Text(
          'مرحباً، $username',
          style: TextStyle(fontSize: 24),
        ),
      ),
    );
  }
}.
